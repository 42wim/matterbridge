package pubsub

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
)

const (
	defaultValidateQueueSize   = 32
	defaultValidateConcurrency = 1024
	defaultValidateThrottle    = 8192
)

// ValidationError is an error that may be signalled from message publication when the message
// fails validation
type ValidationError struct {
	Reason string
}

func (e ValidationError) Error() string {
	return e.Reason
}

// Validator is a function that validates a message with a binary decision: accept or reject.
type Validator func(context.Context, peer.ID, *Message) bool

// ValidatorEx is an extended validation function that validates a message with an enumerated decision
type ValidatorEx func(context.Context, peer.ID, *Message) ValidationResult

// ValidationResult represents the decision of an extended validator
type ValidationResult int

const (
	// ValidationAccept is a validation decision that indicates a valid message that should be accepted and
	// delivered to the application and forwarded to the network.
	ValidationAccept = ValidationResult(0)
	// ValidationReject is a validation decision that indicates an invalid message that should not be
	// delivered to the application or forwarded to the application. Furthermore the peer that forwarded
	// the message should be penalized by peer scoring routers.
	ValidationReject = ValidationResult(1)
	// ValidationIgnore is a validation decision that indicates a message that should be ignored: it will
	// be neither delivered to the application nor forwarded to the network. However, in contrast to
	// ValidationReject, the peer that forwarded the message must not be penalized by peer scoring routers.
	ValidationIgnore = ValidationResult(2)
	// internal
	validationThrottled = ValidationResult(-1)
)

// ValidatorOpt is an option for RegisterTopicValidator.
type ValidatorOpt func(addVal *addValReq) error

// validation represents the validator pipeline.
// The validator pipeline performs signature validation and runs a
// sequence of user-configured validators per-topic. It is possible to
// adjust various concurrency parameters, such as the number of
// workers and the max number of simultaneous validations. The user
// can also attach inline validators that will be executed
// synchronously; this may be useful to prevent superfluous
// context-switching for lightweight tasks.
type validation struct {
	p *PubSub

	tracer *pubsubTracer

	// mx protects the validator map
	mx sync.Mutex
	// topicVals tracks per topic validators
	topicVals map[string]*validatorImpl

	// defaultVals tracks default validators applicable to all topics
	defaultVals []*validatorImpl

	// validateQ is the front-end to the validation pipeline
	validateQ chan *validateReq

	// validateThrottle limits the number of active validation goroutines
	validateThrottle chan struct{}

	// this is the number of synchronous validation workers
	validateWorkers int
}

// validation requests
type validateReq struct {
	vals []*validatorImpl
	src  peer.ID
	msg  *Message
}

// representation of topic validators
type validatorImpl struct {
	topic            string
	validate         ValidatorEx
	validateTimeout  time.Duration
	validateThrottle chan struct{}
	validateInline   bool
}

// async request to add a topic validators
type addValReq struct {
	topic    string
	validate interface{}
	timeout  time.Duration
	throttle int
	inline   bool
	resp     chan error
}

// async request to remove a topic validator
type rmValReq struct {
	topic string
	resp  chan error
}

// newValidation creates a new validation pipeline
func newValidation() *validation {
	return &validation{
		topicVals:        make(map[string]*validatorImpl),
		validateQ:        make(chan *validateReq, defaultValidateQueueSize),
		validateThrottle: make(chan struct{}, defaultValidateThrottle),
		validateWorkers:  runtime.NumCPU(),
	}
}

// Start attaches the validation pipeline to a pubsub instance and starts background
// workers
func (v *validation) Start(p *PubSub) {
	v.p = p
	v.tracer = p.tracer
	for i := 0; i < v.validateWorkers; i++ {
		go v.validateWorker()
	}
}

// AddValidator adds a new validator
func (v *validation) AddValidator(req *addValReq) {
	val, err := v.makeValidator(req)
	if err != nil {
		req.resp <- err
		return
	}

	v.mx.Lock()
	defer v.mx.Unlock()

	topic := val.topic

	_, ok := v.topicVals[topic]
	if ok {
		req.resp <- fmt.Errorf("duplicate validator for topic %s", topic)
		return
	}

	v.topicVals[topic] = val
	req.resp <- nil
}

func (v *validation) makeValidator(req *addValReq) (*validatorImpl, error) {
	makeValidatorEx := func(v Validator) ValidatorEx {
		return func(ctx context.Context, p peer.ID, msg *Message) ValidationResult {
			if v(ctx, p, msg) {
				return ValidationAccept
			} else {
				return ValidationReject
			}
		}
	}

	var validator ValidatorEx
	switch v := req.validate.(type) {
	case func(ctx context.Context, p peer.ID, msg *Message) bool:
		validator = makeValidatorEx(Validator(v))
	case Validator:
		validator = makeValidatorEx(v)

	case func(ctx context.Context, p peer.ID, msg *Message) ValidationResult:
		validator = ValidatorEx(v)
	case ValidatorEx:
		validator = v

	default:
		topic := req.topic
		if req.topic == "" {
			topic = "(default)"
		}
		return nil, fmt.Errorf("unknown validator type for topic %s; must be an instance of Validator or ValidatorEx", topic)
	}

	val := &validatorImpl{
		topic:            req.topic,
		validate:         validator,
		validateTimeout:  0,
		validateThrottle: make(chan struct{}, defaultValidateConcurrency),
		validateInline:   req.inline,
	}

	if req.timeout > 0 {
		val.validateTimeout = req.timeout
	}

	if req.throttle > 0 {
		val.validateThrottle = make(chan struct{}, req.throttle)
	}

	return val, nil
}

// RemoveValidator removes an existing validator
func (v *validation) RemoveValidator(req *rmValReq) {
	v.mx.Lock()
	defer v.mx.Unlock()

	topic := req.topic

	_, ok := v.topicVals[topic]
	if ok {
		delete(v.topicVals, topic)
		req.resp <- nil
	} else {
		req.resp <- fmt.Errorf("no validator for topic %s", topic)
	}
}

// PushLocal synchronously pushes a locally published message and performs applicable
// validations.
// Returns an error if validation fails
func (v *validation) PushLocal(msg *Message) error {
	v.p.tracer.PublishMessage(msg)

	err := v.p.checkSigningPolicy(msg)
	if err != nil {
		return err
	}

	vals := v.getValidators(msg)
	return v.validate(vals, msg.ReceivedFrom, msg, true)
}

// Push pushes a message into the validation pipeline.
// It returns true if the message can be forwarded immediately without validation.
func (v *validation) Push(src peer.ID, msg *Message) bool {
	vals := v.getValidators(msg)

	if len(vals) > 0 || msg.Signature != nil {
		select {
		case v.validateQ <- &validateReq{vals, src, msg}:
		default:
			log.Debugf("message validation throttled: queue full; dropping message from %s", src)
			v.tracer.RejectMessage(msg, RejectValidationQueueFull)
		}
		return false
	}

	return true
}

// getValidators returns all validators that apply to a given message
func (v *validation) getValidators(msg *Message) []*validatorImpl {
	v.mx.Lock()
	defer v.mx.Unlock()

	var vals []*validatorImpl
	vals = append(vals, v.defaultVals...)

	topic := msg.GetTopic()

	val, ok := v.topicVals[topic]
	if !ok {
		return vals
	}

	return append(vals, val)
}

// validateWorker is an active goroutine performing inline validation
func (v *validation) validateWorker() {
	for {
		select {
		case req := <-v.validateQ:
			v.validate(req.vals, req.src, req.msg, false)
		case <-v.p.ctx.Done():
			return
		}
	}
}

// validate performs validation and only sends the message if all validators succeed
func (v *validation) validate(vals []*validatorImpl, src peer.ID, msg *Message, synchronous bool) error {
	// If signature verification is enabled, but signing is disabled,
	// the Signature is required to be nil upon receiving the message in PubSub.pushMsg.
	if msg.Signature != nil {
		if !v.validateSignature(msg) {
			log.Debugf("message signature validation failed; dropping message from %s", src)
			v.tracer.RejectMessage(msg, RejectInvalidSignature)
			return ValidationError{Reason: RejectInvalidSignature}
		}
	}

	// we can mark the message as seen now that we have verified the signature
	// and avoid invoking user validators more than once
	id := v.p.idGen.ID(msg)
	if !v.p.markSeen(id) {
		v.tracer.DuplicateMessage(msg)
		return nil
	} else {
		v.tracer.ValidateMessage(msg)
	}

	var inline, async []*validatorImpl
	for _, val := range vals {
		if val.validateInline || synchronous {
			inline = append(inline, val)
		} else {
			async = append(async, val)
		}
	}

	// apply inline (synchronous) validators
	result := ValidationAccept
loop:
	for _, val := range inline {
		switch val.validateMsg(v.p.ctx, src, msg) {
		case ValidationAccept:
		case ValidationReject:
			result = ValidationReject
			break loop
		case ValidationIgnore:
			result = ValidationIgnore
		}
	}

	if result == ValidationReject {
		log.Debugf("message validation failed; dropping message from %s", src)
		v.tracer.RejectMessage(msg, RejectValidationFailed)
		return ValidationError{Reason: RejectValidationFailed}
	}

	// apply async validators
	if len(async) > 0 {
		select {
		case v.validateThrottle <- struct{}{}:
			go func() {
				v.doValidateTopic(async, src, msg, result)
				<-v.validateThrottle
			}()
		default:
			log.Debugf("message validation throttled; dropping message from %s", src)
			v.tracer.RejectMessage(msg, RejectValidationThrottled)
		}
		return nil
	}

	if result == ValidationIgnore {
		v.tracer.RejectMessage(msg, RejectValidationIgnored)
		return ValidationError{Reason: RejectValidationIgnored}
	}

	// no async validators, accepted message, send it!
	select {
	case v.p.sendMsg <- msg:
		return nil
	case <-v.p.ctx.Done():
		return v.p.ctx.Err()
	}
}

func (v *validation) validateSignature(msg *Message) bool {
	err := verifyMessageSignature(msg.Message)
	if err != nil {
		log.Debugf("signature verification error: %s", err.Error())
		return false
	}

	return true
}

func (v *validation) doValidateTopic(vals []*validatorImpl, src peer.ID, msg *Message, r ValidationResult) {
	result := v.validateTopic(vals, src, msg)

	if result == ValidationAccept && r != ValidationAccept {
		result = r
	}

	switch result {
	case ValidationAccept:
		v.p.sendMsg <- msg
	case ValidationReject:
		log.Debugf("message validation failed; dropping message from %s", src)
		v.tracer.RejectMessage(msg, RejectValidationFailed)
		return
	case ValidationIgnore:
		log.Debugf("message validation punted; ignoring message from %s", src)
		v.tracer.RejectMessage(msg, RejectValidationIgnored)
		return
	case validationThrottled:
		log.Debugf("message validation throttled; ignoring message from %s", src)
		v.tracer.RejectMessage(msg, RejectValidationThrottled)

	default:
		// BUG: this would be an internal programming error, so a panic seems appropiate.
		panic(fmt.Errorf("unexpected validation result: %d", result))
	}
}

func (v *validation) validateTopic(vals []*validatorImpl, src peer.ID, msg *Message) ValidationResult {
	if len(vals) == 1 {
		return v.validateSingleTopic(vals[0], src, msg)
	}

	ctx, cancel := context.WithCancel(v.p.ctx)
	defer cancel()

	rch := make(chan ValidationResult, len(vals))
	rcount := 0

	for _, val := range vals {
		rcount++

		select {
		case val.validateThrottle <- struct{}{}:
			go func(val *validatorImpl) {
				rch <- val.validateMsg(ctx, src, msg)
				<-val.validateThrottle
			}(val)

		default:
			log.Debugf("validation throttled for topic %s", val.topic)
			rch <- validationThrottled
		}
	}

	result := ValidationAccept
loop:
	for i := 0; i < rcount; i++ {
		switch <-rch {
		case ValidationAccept:
		case ValidationReject:
			result = ValidationReject
			break loop
		case ValidationIgnore:
			// throttled validation has the same effect, but takes precedence over Ignore as it is not
			// known whether the throttled validator would have signaled rejection.
			if result != validationThrottled {
				result = ValidationIgnore
			}
		case validationThrottled:
			result = validationThrottled
		}
	}

	return result
}

// fast path for single topic validation that avoids the extra goroutine
func (v *validation) validateSingleTopic(val *validatorImpl, src peer.ID, msg *Message) ValidationResult {
	select {
	case val.validateThrottle <- struct{}{}:
		res := val.validateMsg(v.p.ctx, src, msg)
		<-val.validateThrottle
		return res

	default:
		log.Debugf("validation throttled for topic %s", val.topic)
		return validationThrottled
	}
}

func (val *validatorImpl) validateMsg(ctx context.Context, src peer.ID, msg *Message) ValidationResult {
	start := time.Now()
	defer func() {
		log.Debugf("validation done; took %s", time.Since(start))
	}()

	if val.validateTimeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, val.validateTimeout)
		defer cancel()
	}

	r := val.validate(ctx, src, msg)
	switch r {
	case ValidationAccept:
		fallthrough
	case ValidationReject:
		fallthrough
	case ValidationIgnore:
		return r

	default:
		log.Warnf("Unexpected result from validator: %d; ignoring message", r)
		return ValidationIgnore
	}
}

// / Options
// WithDefaultValidator adds a validator that applies to all topics by default; it can be used
// more than once and add multiple validators. Having a defult validator does not inhibit registering
// a per topic validator.
func WithDefaultValidator(val interface{}, opts ...ValidatorOpt) Option {
	return func(ps *PubSub) error {
		addVal := &addValReq{
			validate: val,
		}

		for _, opt := range opts {
			err := opt(addVal)
			if err != nil {
				return err
			}
		}

		val, err := ps.val.makeValidator(addVal)
		if err != nil {
			return err
		}

		ps.val.defaultVals = append(ps.val.defaultVals, val)
		return nil
	}
}

// WithValidateQueueSize sets the buffer of validate queue. Defaults to 32.
// When queue is full, validation is throttled and new messages are dropped.
func WithValidateQueueSize(n int) Option {
	return func(ps *PubSub) error {
		if n > 0 {
			ps.val.validateQ = make(chan *validateReq, n)
			return nil
		}
		return fmt.Errorf("validate queue size must be > 0")
	}
}

// WithValidateThrottle sets the upper bound on the number of active validation
// goroutines across all topics. The default is 8192.
func WithValidateThrottle(n int) Option {
	return func(ps *PubSub) error {
		ps.val.validateThrottle = make(chan struct{}, n)
		return nil
	}
}

// WithValidateWorkers sets the number of synchronous validation worker goroutines.
// Defaults to NumCPU.
//
// The synchronous validation workers perform signature validation, apply inline
// user validators, and schedule asynchronous user validators.
// You can adjust this parameter to devote less cpu time to synchronous validation.
func WithValidateWorkers(n int) Option {
	return func(ps *PubSub) error {
		if n > 0 {
			ps.val.validateWorkers = n
			return nil
		}
		return fmt.Errorf("number of validation workers must be > 0")
	}
}

// WithValidatorTimeout is an option that sets a timeout for an (asynchronous) topic validator.
// By default there is no timeout in asynchronous validators.
func WithValidatorTimeout(timeout time.Duration) ValidatorOpt {
	return func(addVal *addValReq) error {
		addVal.timeout = timeout
		return nil
	}
}

// WithValidatorConcurrency is an option that sets the topic validator throttle.
// This controls the number of active validation goroutines for the topic; the default is 1024.
func WithValidatorConcurrency(n int) ValidatorOpt {
	return func(addVal *addValReq) error {
		addVal.throttle = n
		return nil
	}
}

// WithValidatorInline is an option that sets the validation disposition to synchronous:
// it will be executed inline in validation front-end, without spawning a new goroutine.
// This is suitable for simple or cpu-bound validators that do not block.
func WithValidatorInline(inline bool) ValidatorOpt {
	return func(addVal *addValReq) error {
		addVal.inline = inline
		return nil
	}
}
