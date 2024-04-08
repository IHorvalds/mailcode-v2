# Mailcodes v2

*Once again, v1 was a write-off...*

Watch mailboxes using IMAP IDLE or polling with the [go-imap](https://github.com/emersion/go-imap) library.

Currently, there is a service and a control cmd. The service loads a config that looks like this:
```yaml
database: /Users/horvalds/Development/mailcodes/service/build/emails.db
subjects:
  - verify
  - verification
  - auth
extractors:
  - regex: ">\\s*(?<code>\\d{4,8})\\s*<"
    capture: "code"
  - regex: "LinkedIn account\\.\\r\\n\\r\\n(?<code>\\d{6})"
    capture: "code"
```

- The "database" is the path where the credentials for the mailboxes are stored.
- The "subjects" list contains the subjects that the service will search all mailboxes for. Case insensitive.
- The "extractors" are tuples (capturing regex, capture group index/name) for extracting authentication codes. For each new email with one of the subjects in its "subject" field, each one of the extractors will be applied to the email's body until one has a match or none remain.

An extractor can capture either by index or by name. E.g.
```yaml
extractors:
  # by index. Captures $1
  - regex: ">\\s*(\\d{4,8})\\s*<"
    capture: 1

    # by name. Captures "code" group
  - regex: ">\\s*(?<code>\\d{4,8})\\s*<"
    capture: "code"
```

## TODOs
- [ ] Add unit tests for config loading, parsing, message parsing, message handling.
- [ ] Add UI for Mac. Needs to be able to send and receive messages over unix sockets.
- [ ] Add communication over named pipes on windows
- [ ] Add UI for Windows.
- [ ] (Eventually) Add UI for Linux.
