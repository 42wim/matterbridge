package protocol

import (
	"sync"

	"go.uber.org/zap"

	"github.com/status-im/status-go/protocol/encryption/multidevice"
	"github.com/status-im/status-go/protocol/protobuf"
)

/*
|--------------------------------------------------------------------------
| chatMap
|--------------------------------------------------------------------------
|
| A sync.Map wrapper for a specific mapping of map[string]*Chat
|
*/

type chatMap struct {
	sm sync.Map
}

func (cm *chatMap) Load(chatID string) (*Chat, bool) {
	chat, ok := cm.sm.Load(chatID)
	if chat == nil {
		return nil, ok
	}
	return chat.(*Chat), ok
}

func (cm *chatMap) Store(chatID string, chat *Chat) {
	cm.sm.Store(chatID, chat)
}

func (cm *chatMap) Range(f func(chatID string, chat *Chat) (shouldContinue bool)) {
	nf := func(key, value interface{}) (shouldContinue bool) {
		return f(key.(string), value.(*Chat))
	}
	cm.sm.Range(nf)
}

func (cm *chatMap) Delete(chatID string) {
	cm.sm.Delete(chatID)
}

/*
|--------------------------------------------------------------------------
| contactMap
|--------------------------------------------------------------------------
|
| A sync.Map wrapper for a specific mapping of map[string]*Contact
|
*/

type contactMap struct {
	sm     sync.Map
	me     *Contact
	logger *zap.Logger
}

func (cm *contactMap) Load(contactID string) (*Contact, bool) {
	if contactID == cm.me.ID {
		cm.logger.Warn("contacts map: loading own identity", zap.String("contactID", contactID))
		return cm.me, true
	}
	contact, ok := cm.sm.Load(contactID)
	if contact == nil {
		return nil, ok
	}
	return contact.(*Contact), ok
}

func (cm *contactMap) Store(contactID string, contact *Contact) {
	if contactID == cm.me.ID {
		cm.logger.Warn("contacts map: storing own identity", zap.String("contactID", contactID))
		return
	}
	cm.sm.Store(contactID, contact)
}

func (cm *contactMap) Range(f func(contactID string, contact *Contact) (shouldContinue bool)) {
	nf := func(key, value interface{}) (shouldContinue bool) {
		return f(key.(string), value.(*Contact))
	}
	cm.sm.Range(nf)
}

func (cm *contactMap) Delete(contactID string) {
	if contactID == cm.me.ID {
		cm.logger.Warn("contacts map: deleting own identity", zap.String("contactID", contactID))
		return
	}
	cm.sm.Delete(contactID)
}

func (cm *contactMap) Len() int {
	count := 0
	cm.Range(func(contactID string, contact *Contact) (shouldContinue bool) {
		count++
		return true
	})

	return count
}

/*
|--------------------------------------------------------------------------
| systemMessageTranslationsMap
|--------------------------------------------------------------------------
|
| A sync.Map wrapper for the specific mapping of map[protobuf.MembershipUpdateEvent_EventType]string
|
*/

type systemMessageTranslationsMap struct {
	sm sync.Map
}

func (smtm *systemMessageTranslationsMap) Init(set map[protobuf.MembershipUpdateEvent_EventType]string) {
	for eventType, message := range set {
		smtm.Store(eventType, message)
	}
}

func (smtm *systemMessageTranslationsMap) Load(eventType protobuf.MembershipUpdateEvent_EventType) (string, bool) {
	message, ok := smtm.sm.Load(eventType)
	if message == nil {
		return "", ok
	}
	return message.(string), ok
}

func (smtm *systemMessageTranslationsMap) Store(eventType protobuf.MembershipUpdateEvent_EventType, message string) {
	smtm.sm.Store(eventType, message)
}

func (smtm *systemMessageTranslationsMap) Range(f func(eventType protobuf.MembershipUpdateEvent_EventType, message string) (shouldContinue bool)) {
	nf := func(key, value interface{}) (shouldContinue bool) {
		return f(key.(protobuf.MembershipUpdateEvent_EventType), value.(string))
	}
	smtm.sm.Range(nf)
}

func (smtm *systemMessageTranslationsMap) Delete(eventType protobuf.MembershipUpdateEvent_EventType) {
	smtm.sm.Delete(eventType)
}

/*
|--------------------------------------------------------------------------
| installationMap
|--------------------------------------------------------------------------
|
| A sync.Map wrapper for the specific mapping of map[string]*multidevice.Installation
|
*/

type installationMap struct {
	sm sync.Map
}

func (im *installationMap) Load(installationID string) (*multidevice.Installation, bool) {
	installation, ok := im.sm.Load(installationID)
	if installation == nil {
		return nil, ok
	}
	return installation.(*multidevice.Installation), ok
}

func (im *installationMap) Store(installationID string, installation *multidevice.Installation) {
	im.sm.Store(installationID, installation)
}

func (im *installationMap) Range(f func(installationID string, installation *multidevice.Installation) (shouldContinue bool)) {
	nf := func(key, value interface{}) (shouldContinue bool) {
		return f(key.(string), value.(*multidevice.Installation))
	}
	im.sm.Range(nf)
}

func (im *installationMap) Delete(installationID string) {
	im.sm.Delete(installationID)
}

func (im *installationMap) Empty() bool {
	count := 0
	im.Range(func(installationID string, installation *multidevice.Installation) (shouldContinue bool) {
		count++
		return false
	})

	return count == 0
}

func (im *installationMap) Len() int {
	count := 0
	im.Range(func(installationID string, installation *multidevice.Installation) (shouldContinue bool) {
		count++
		return true
	})

	return count
}

/*
|--------------------------------------------------------------------------
| stringBoolMap
|--------------------------------------------------------------------------
|
| A sync.Map wrapper for the specific mapping of map[string]bool
|
*/

type stringBoolMap struct {
	sm sync.Map
}

func (sbm *stringBoolMap) Load(key string) (bool, bool) {
	state, ok := sbm.sm.Load(key)
	if state == nil {
		return false, ok
	}
	return state.(bool), ok
}

func (sbm *stringBoolMap) Store(key string, value bool) {
	sbm.sm.Store(key, value)
}

func (sbm *stringBoolMap) Range(f func(key string, value bool) (shouldContinue bool)) {
	nf := func(key, value interface{}) (shouldContinue bool) {
		return f(key.(string), value.(bool))
	}
	sbm.sm.Range(nf)
}

func (sbm *stringBoolMap) Delete(key string) {
	sbm.sm.Delete(key)
}

func (sbm *stringBoolMap) Len() int {
	count := 0
	sbm.Range(func(key string, value bool) (shouldContinue bool) {
		count++
		return true
	})

	return count
}
