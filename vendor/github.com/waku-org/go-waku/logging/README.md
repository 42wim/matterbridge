# Logging Style Guide

The goal of the style described here is to yield logs that are amenable to searching and aggregating. Structured logging is the best foundation for that. The log entries should be consistent and predictable to support search efficiency and high fidelity of search results. This style puts forward guidelines that promote this outcome.

## Log messages

* Messages should be fixed strings, never interpolate values into the messages. Use log entry fields for values.

* Message strings should be consistent identification of what action/event was/is happening. Consistent messages makes searching the logs and aggregating correlated events easier.

* Error messages should look like any other log messages. No need to say "x failed", the log level and error field are sufficient indication of failure.

## Log entry fields

* Adding fields to log entries is not free, but the fields are the discriminators that allow distinguishing similar log entries from each other. Insufficient field structure will makes it more difficult to find the entries you are looking for.

* Create/Use field helpers for commonly used field value types (see logging.go). It promotes consistent formatting and allows changing it easily in one place.

# Log entry field helpers

* Make the field creation do as little as possible, i.e. just capture the existing value/object. Postpone any transformation to log emission time by employing generic zap.Stringer, zap.Array, zap.Object fields (see logging.go). It avoids unnecessary transformation for entries that may not even be emitted in the end.

## Logger management

* Adorn the logger with fields and reuse the adorned logger rather than repeatedly creating fields with each log entry.

* Prefer passing the adorned logger down the call chain using Context. It promotes consistent log entry structure, i.e. fields will exist consistently in related entries.
