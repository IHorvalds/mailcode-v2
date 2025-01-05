# Mailcodes v2

*Once again, v1 was a write-off...*

Watch mailboxes using IMAP IDLE or polling with the [go-imap](https://github.com/emersion/go-imap) library.

# Building

> [!IMPORTANT]
Make sure you have a file called `google-oauth-vars.mk` next to the Makefile
It must contain the following as they are compiled into the binary:
```Makefile
GOOGLE_CLIENT_ID := ${Your app's google client ID}
GOOGLE_CLIENT_SECRET := ${Your app's google client secret}
```

> [!IMPORTANT]
Make sure your Google App (in the developer dashboard) has the following permissions:
- "email"
- "https://mail.google.com/"

## For development
This should download the dependencies and start a "live" reloading build with air and templ --watch
```shell
make live
```

## For "release"

```shell
make build
```