package kbchat

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/keybase1"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/stellar1"
)

// SubscriptionMessage contains a message and conversation object
type SubscriptionMessage struct {
	Message      chat1.MsgSummary
	Conversation chat1.ConvSummary
}

type SubscriptionConversation struct {
	Conversation chat1.ConvSummary
}

type SubscriptionWalletEvent struct {
	Payment stellar1.PaymentDetailsLocal
}

// Subscription has methods to control the background message fetcher loop
type Subscription struct {
	*DebugOutput
	sync.Mutex

	newMsgsCh   chan SubscriptionMessage
	newConvsCh  chan SubscriptionConversation
	newWalletCh chan SubscriptionWalletEvent
	errorCh     chan error
	running     bool
	shutdownCh  chan struct{}
}

func NewSubscription() *Subscription {
	newMsgsCh := make(chan SubscriptionMessage, 100)
	newConvsCh := make(chan SubscriptionConversation, 100)
	newWalletCh := make(chan SubscriptionWalletEvent, 100)
	errorCh := make(chan error, 100)
	shutdownCh := make(chan struct{})
	return &Subscription{
		DebugOutput: NewDebugOutput("Subscription"),
		newMsgsCh:   newMsgsCh,
		newConvsCh:  newConvsCh,
		newWalletCh: newWalletCh,
		shutdownCh:  shutdownCh,
		errorCh:     errorCh,
		running:     true,
	}
}

// Read blocks until a new message arrives
func (m *Subscription) Read() (msg SubscriptionMessage, err error) {
	defer m.Trace(&err, "Read")()
	select {
	case msg = <-m.newMsgsCh:
		return msg, nil
	case err = <-m.errorCh:
		return SubscriptionMessage{}, err
	case <-m.shutdownCh:
		return SubscriptionMessage{}, errors.New("Subscription shutdown")
	}
}

func (m *Subscription) ReadNewConvs() (conv SubscriptionConversation, err error) {
	defer m.Trace(&err, "ReadNewConvs")()
	select {
	case conv = <-m.newConvsCh:
		return conv, nil
	case err = <-m.errorCh:
		return SubscriptionConversation{}, err
	case <-m.shutdownCh:
		return SubscriptionConversation{}, errors.New("Subscription shutdown")
	}
}

// Read blocks until a new message arrives
func (m *Subscription) ReadWallet() (msg SubscriptionWalletEvent, err error) {
	defer m.Trace(&err, "ReadWallet")()
	select {
	case msg = <-m.newWalletCh:
		return msg, nil
	case err = <-m.errorCh:
		return SubscriptionWalletEvent{}, err
	case <-m.shutdownCh:
		return SubscriptionWalletEvent{}, errors.New("Subscription shutdown")
	}
}

// Shutdown terminates the background process
func (m *Subscription) Shutdown() {
	defer m.Trace(nil, "Shutdown")()
	m.Lock()
	defer m.Unlock()
	if m.running {
		close(m.shutdownCh)
		m.running = false
	}
}

type ListenOptions struct {
	Wallet bool
	Convs  bool
}

type PaymentHolder struct {
	Payment stellar1.PaymentDetailsLocal `json:"notification"`
}

type TypeHolder struct {
	Type string `json:"type"`
}

type OneshotOptions struct {
	Username string
	PaperKey string
}

type RunOptions struct {
	KeybaseLocation string
	HomeDir         string
	Oneshot         *OneshotOptions
	StartService    bool
	// Have the bot send/receive typing notifications
	EnableTyping bool
	// Disable bot lite mode
	DisableBotLiteMode bool
}

func (r RunOptions) Location() string {
	if r.KeybaseLocation == "" {
		return "keybase"
	}
	return r.KeybaseLocation
}

func (r RunOptions) Command(args ...string) *exec.Cmd {
	var cmd []string
	if r.HomeDir != "" {
		cmd = append(cmd, "--home", r.HomeDir)
	}
	cmd = append(cmd, args...)
	return exec.Command(r.Location(), cmd...)
}

// Start fires up the Keybase JSON API in stdin/stdout mode
func Start(runOpts RunOptions, opts ...func(*API)) (*API, error) {
	api := NewAPI(runOpts, opts...)
	if err := api.startPipes(); err != nil {
		return nil, err
	}
	return api, nil
}

// API is the main object used for communicating with the Keybase JSON API
type API struct {
	sync.Mutex
	*DebugOutput
	apiInput      io.Writer
	apiOutput     *bufio.Reader
	apiCmd        *exec.Cmd
	username      string
	runOpts       RunOptions
	subscriptions []*Subscription
	Timeout       time.Duration
	LogSendBytes  int
}

func CustomTimeout(timeout time.Duration) func(*API) {
	return func(a *API) {
		a.Timeout = timeout
	}
}

func NewAPI(runOpts RunOptions, opts ...func(*API)) *API {
	api := &API{
		DebugOutput:  NewDebugOutput("API"),
		runOpts:      runOpts,
		Timeout:      5 * time.Second,
		LogSendBytes: 1024 * 1024 * 5, // request 5MB so we don't get killed
	}
	for _, opt := range opts {
		opt(api)
	}
	return api
}

func (a *API) Command(args ...string) *exec.Cmd {
	return a.runOpts.Command(args...)
}

func (a *API) getUsername(runOpts RunOptions) (username string, err error) {
	p := runOpts.Command("whoami", "-json")
	output, err := p.StdoutPipe()
	if err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" {
		p.ExtraFiles = []*os.File{output.(*os.File)}
	}
	if err = p.Start(); err != nil {
		return "", err
	}

	doneCh := make(chan error)
	go func() {
		defer func() { close(doneCh) }()
		statusJSON, err := ioutil.ReadAll(output)
		if err != nil {
			doneCh <- fmt.Errorf("error reading whoami output: %v", err)
			return
		}
		var status keybase1.CurrentStatus
		if err := json.Unmarshal(statusJSON, &status); err != nil {
			doneCh <- fmt.Errorf("invalid whoami JSON %q: %v", statusJSON, err)
			return
		}
		if status.LoggedIn && status.User != nil {
			username = status.User.Username
			doneCh <- nil
		} else {
			doneCh <- fmt.Errorf("unable to authenticate to keybase service: logged in: %v user: %+v", status.LoggedIn, status.User)
		}
		// Cleanup the command
		if err := p.Wait(); err != nil {
			a.Debug("unable to wait for cmd: %v", err)
		}
	}()

	select {
	case err = <-doneCh:
		if err != nil {
			return "", err
		}
	case <-time.After(a.Timeout):
		return "", errors.New("unable to run Keybase command")
	}

	return username, nil
}

func (a *API) auth() (string, error) {
	username, err := a.getUsername(a.runOpts)
	if err == nil {
		return username, nil
	}
	if a.runOpts.Oneshot == nil {
		return "", err
	}
	username = ""
	// If a paper key is specified, then login with oneshot mode (logout first)
	if a.runOpts.Oneshot != nil {
		if username == a.runOpts.Oneshot.Username {
			// just get out if we are on the desired user already
			return username, nil
		}
		if err := a.runOpts.Command("logout", "-f").Run(); err != nil {
			return "", err
		}
		if err := a.runOpts.Command("oneshot", "--username", a.runOpts.Oneshot.Username, "--paperkey",
			a.runOpts.Oneshot.PaperKey).Run(); err != nil {
			return "", err
		}
		username = a.runOpts.Oneshot.Username
		return username, nil
	}
	return "", errors.New("unable to auth")
}

func (a *API) startPipes() (err error) {
	a.Lock()
	defer a.Unlock()
	if a.apiCmd != nil {
		if err := a.apiCmd.Process.Kill(); err != nil {
			return fmt.Errorf("unable to kill previous API command %v", err)
		}
	}
	a.apiCmd = nil

	if a.runOpts.StartService {
		args := []string{fmt.Sprintf("-enable-bot-lite-mode=%v", a.runOpts.DisableBotLiteMode), "service"}
		if err := a.runOpts.Command(args...).Start(); err != nil {
			return fmt.Errorf("unable to start service %v", err)
		}
	}

	if a.username, err = a.auth(); err != nil {
		return fmt.Errorf("unable to auth: %v", err)
	}

	cmd := a.runOpts.Command("chat", "notification-settings", fmt.Sprintf("-disable-typing=%v", !a.runOpts.EnableTyping))
	if err = cmd.Run(); err != nil {
		// This is a performance optimization but isn't a fatal error.
		a.Debug("unable to set notifiation settings %v", err)
	}

	a.apiCmd = a.runOpts.Command("chat", "api")
	if a.apiInput, err = a.apiCmd.StdinPipe(); err != nil {
		return fmt.Errorf("unable to get api stdin: %v", err)
	}
	output, err := a.apiCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("unable to get api stdout: %v", err)
	}
	if runtime.GOOS != "windows" {
		a.apiCmd.ExtraFiles = []*os.File{output.(*os.File)}
	}
	if err := a.apiCmd.Start(); err != nil {
		return fmt.Errorf("unable to run chat api cmd: %v", err)
	}
	a.apiOutput = bufio.NewReader(output)
	return nil
}

func (a *API) getAPIPipesLocked() (io.Writer, *bufio.Reader, error) {
	// this should only be called inside a lock
	if a.apiCmd == nil {
		return nil, nil, errAPIDisconnected
	}
	return a.apiInput, a.apiOutput, nil
}

func (a *API) GetUsername() string {
	return a.username
}

func (a *API) doSend(arg interface{}) (resp SendResponse, err error) {
	a.Lock()
	defer a.Unlock()

	bArg, err := json.Marshal(arg)
	if err != nil {
		return SendResponse{}, fmt.Errorf("unable to send arg: %+v: %v", arg, err)
	}
	input, output, err := a.getAPIPipesLocked()
	if err != nil {
		return SendResponse{}, err
	}
	if _, err := io.WriteString(input, string(bArg)); err != nil {
		return SendResponse{}, err
	}
	responseRaw, err := output.ReadBytes('\n')
	if err != nil {
		return SendResponse{}, err
	}
	if err := json.Unmarshal(responseRaw, &resp); err != nil {
		return resp, fmt.Errorf("failed to decode API response: %v %v", responseRaw, err)
	} else if resp.Error != nil {
		return resp, errors.New(resp.Error.Message)
	}
	return resp, nil
}

func (a *API) doFetch(apiInput string) ([]byte, error) {
	a.Lock()
	defer a.Unlock()

	input, output, err := a.getAPIPipesLocked()
	if err != nil {
		return nil, err
	}
	if _, err := io.WriteString(input, apiInput); err != nil {
		return nil, err
	}
	byteOutput, err := output.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	return byteOutput, nil
}

// ListenForNewTextMessages proxies to Listen without wallet events
func (a *API) ListenForNewTextMessages() (*Subscription, error) {
	opts := ListenOptions{Wallet: false}
	return a.Listen(opts)
}

func (a *API) registerSubscription(sub *Subscription) {
	a.Lock()
	defer a.Unlock()
	a.subscriptions = append(a.subscriptions, sub)
}

// Listen fires of a background loop and puts chat messages and wallet
// events into channels
func (a *API) Listen(opts ListenOptions) (*Subscription, error) {
	done := make(chan struct{})
	sub := NewSubscription()
	a.registerSubscription(sub)
	pause := 2 * time.Second
	readScanner := func(boutput *bufio.Scanner) {
		defer func() { done <- struct{}{} }()
		for {
			select {
			case <-sub.shutdownCh:
				a.Debug("readScanner: received shutdown")
				return
			default:
			}
			boutput.Scan()
			t := boutput.Text()
			var typeHolder TypeHolder
			if err := json.Unmarshal([]byte(t), &typeHolder); err != nil {
				sub.errorCh <- fmt.Errorf("err: %v, data: %v", err, t)
				break
			}
			switch typeHolder.Type {
			case "chat":
				var notification chat1.MsgNotification
				if err := json.Unmarshal([]byte(t), &notification); err != nil {
					sub.errorCh <- fmt.Errorf("err: %v, data: %v", err, t)
					break
				}
				if notification.Error != nil {
					a.Debug("error message received: %s", *notification.Error)
				} else if notification.Msg != nil {
					subscriptionMessage := SubscriptionMessage{
						Message: *notification.Msg,
						Conversation: chat1.ConvSummary{
							Id:      notification.Msg.ConvID,
							Channel: notification.Msg.Channel,
						},
					}
					sub.newMsgsCh <- subscriptionMessage
				}
			case "chat_conv":
				var notification chat1.ConvNotification
				if err := json.Unmarshal([]byte(t), &notification); err != nil {
					sub.errorCh <- fmt.Errorf("err: %v, data: %v", err, t)
					break
				}
				if notification.Error != nil {
					a.Debug("error message received: %s", *notification.Error)
				} else if notification.Conv != nil {
					subscriptionConv := SubscriptionConversation{
						Conversation: *notification.Conv,
					}
					sub.newConvsCh <- subscriptionConv
				}
			case "wallet":
				var holder PaymentHolder
				if err := json.Unmarshal([]byte(t), &holder); err != nil {
					sub.errorCh <- fmt.Errorf("err: %v, data: %v", err, t)
					break
				}
				subscriptionPayment := SubscriptionWalletEvent(holder)
				sub.newWalletCh <- subscriptionPayment
			default:
				continue
			}
		}
	}

	attempts := 0
	maxAttempts := 30
	go func() {
		defer func() {
			close(sub.newMsgsCh)
			close(sub.newConvsCh)
			close(sub.newWalletCh)
			close(sub.errorCh)
		}()
		for {
			select {
			case <-sub.shutdownCh:
				a.Debug("Listen: received shutdown")
				return
			default:
			}

			if attempts >= maxAttempts {
				if err := a.LogSend("Listen: failed to auth, giving up"); err != nil {
					a.Debug("Listen: logsend failed to send: %v", err)
				}
				panic("Listen: failed to auth, giving up")
			}
			attempts++
			if _, err := a.auth(); err != nil {
				a.Debug("Listen: failed to auth: %s", err)
				time.Sleep(pause)
				continue
			}
			cmdElements := []string{"chat", "api-listen"}
			if opts.Wallet {
				cmdElements = append(cmdElements, "--wallet")
			}
			if opts.Convs {
				cmdElements = append(cmdElements, "--convs")
			}
			p := a.runOpts.Command(cmdElements...)
			output, err := p.StdoutPipe()
			if err != nil {
				a.Debug("Listen: failed to listen: %s", err)
				time.Sleep(pause)
				continue
			}
			stderr, err := p.StderrPipe()
			if err != nil {
				a.Debug("Listen: failed to listen to stderr: %s", err)
				time.Sleep(pause)
				continue
			}
			if runtime.GOOS != "windows" {
				p.ExtraFiles = []*os.File{stderr.(*os.File), output.(*os.File)}
			}
			boutput := bufio.NewScanner(output)
			if err := p.Start(); err != nil {

				a.Debug("Listen: failed to make listen scanner: %s", err)
				time.Sleep(pause)
				continue
			}
			attempts = 0
			go readScanner(boutput)
			select {
			case <-sub.shutdownCh:
				a.Debug("Listen: received shutdown")
				return
			case <-done:
			}
			if err := p.Wait(); err != nil {
				stderrBytes, rerr := ioutil.ReadAll(stderr)
				if rerr != nil {
					stderrBytes = []byte(fmt.Sprintf("failed to get stderr: %v", rerr))
				}
				a.Debug("Listen: failed to Wait for command, restarting pipes: %s (```%s```)", err, stderrBytes)
				if err := a.startPipes(); err != nil {
					a.Debug("Listen: failed to restart pipes: %v", err)
				}
			}
			time.Sleep(pause)
		}
	}()
	return sub, nil
}

func (a *API) LogSend(feedback string) error {
	feedback = "go-keybase-chat-bot log send\n" +
		"username: " + a.GetUsername() + "\n" +
		feedback

	args := []string{
		"log", "send",
		"--no-confirm",
		"--feedback", feedback,
		"-n", fmt.Sprintf("%d", a.LogSendBytes),
	}
	return a.runOpts.Command(args...).Run()
}

func (a *API) Shutdown() (err error) {
	defer a.Trace(&err, "Shutdown")()
	a.Lock()
	defer a.Unlock()
	for _, sub := range a.subscriptions {
		sub.Shutdown()
	}
	if a.apiCmd != nil {
		a.Debug("waiting for API command")
		if err := a.apiCmd.Wait(); err != nil {
			return err
		}
	}

	if a.runOpts.Oneshot != nil {
		a.Debug("logging out")
		err := a.runOpts.Command("logout", "--force").Run()
		if err != nil {
			return err
		}
	}

	if a.runOpts.StartService {
		a.Debug("stopping service")
		err := a.runOpts.Command("ctl", "stop", "--shutdown").Run()
		if err != nil {
			return err
		}
	}

	return nil
}
