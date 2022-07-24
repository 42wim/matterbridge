# maulogger
A logger in Go. Deprecated in favor of [zerolog](https://github.com/rs/zerolog).

Utilities for migrating gracefully can be found in the maulogadapt package,
it includes both wrapping a zerolog in the maulogger interface, and wrapping a
maulogger as a zerolog output writer.
