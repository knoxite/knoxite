# Mega

This is the storage backend for the [mega](https://mega.nz) cloud storage.

# Usage

To store data on mega you need a mega account and its e-mail address and password.
You can use the knoxite URL scheme in order to interact with this backend.
Currently the e-mail address needs to be url encoded.


```
knoxite repo init -r mega://example%40knoxite.com:password@/desired/path
```