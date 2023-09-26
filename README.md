# jamf-prestage-fixup

So someday this might happen where a quick click causes all of your devices in Apple Business or School Manager to become unassigned or assigned to the wrong server.

That's not terribly hard to recover from if you get the device assignment log CSV and split it into 1,000 device lists and paste paste paste.

But if you were using different PreStage Assignment scopes in Jamf, all of those scopings go right out the window.

I wrote this utility to true up Jamf mobile device[^1] prestages, by looking at the last prestage the device enrolled in with and aligning the prestage scope to match.

This is not good code, it was done in just a few hours.

You will need to change the server URL and get a bearer token, paste that into `main.go` and then run it with `go run .`

[^1]: We have very few Mac prestage scopes that aren't exposed as a "MDM Server" to AxM, but the general shape of the API is the same as the mobile devices API. It would not be a lot of work to do Mac or both.
