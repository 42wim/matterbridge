## 2022-09-12 (v1.1.0)

* Create Go module.
* Update examples.
* Fix concurrent panic (PR-65).
* Add `Sessions` to get all sessions (PR-53).
* Add `LocalAddr` and `RemoteAddr` (PR-55).

## 2017-05-18

* Fix `HandleSentMessageBinary`.

## 2017-04-11

* Allow any origin by default.
* Add `BroadcastMultiple`.

## 2017-04-09

* Add control message support.
* Add `IsClosed` to Session.

## 2017-02-10

* Return errors for some exposed methods.
* Add `HandleRequestWithKeys`.
* Add `HandleSentMessage` and `HandleSentMessageBinary`.

## 2017-01-20

* Add `Len()` to fetch number of connected sessions.

## 2016-12-09

* Add metadata management for sessions.

## 2016-05-09

* Add method `HandlePong` to melody instance.

## 2015-10-07

* Add broadcast methods for binary messages.

## 2015-09-03

* Add `Close` method to melody instance.

### 2015-06-10

* Support for binary messages.
* BroadcastOthers method.
