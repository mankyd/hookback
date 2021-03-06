# Hookback

[![Build Status](https://travis-ci.org/mankyd/hookback.svg?branch=master)](https://travis-ci.org/mankyd/hookback)

A simple GitHub WebHook server.

Provide a configuration file. Tell it what command to run for what events and on
which repositories. That's it.

A configuration typically looks like the following:

    {
      "server": {
        "interface": "0.0.0.0",
        "port": "3015"
      },
      "repos": {
        "mankyd/hookback": {
          "secret": "asdf",
          "events":  {
            "ping": {
              "wait": true,
              "cmd": ["echo", "pong"]
            }
          }
        }
      }
    }

If you're familiar with [GitHub's Webhooks](https://developer.github.com/webhooks/),
the breakdown is fairly self explanatory. After enabling webhooks on the
"mankyd/hookback" repository, this server will run "echo pinged" anytime it
receives a "ping" event from GitHub. It expects the webhook secret to be set to
"asdf" and will verify this, rejecting any requests that are not properly signed
with this secret. Secret is not required but is recommended.

Server Configuration
--------------------
The server section has only two options, both shown. By default, the server
listens on all interfaces ("0.0.0.0") and port "3015". If this is acceptable,
you are free to remove the "server" section from the config entirely.

JSON Payload
------------
The JSON payload received from GitHub will be substituted in for "%p" when it
occurs as an argument of the command. If you need "%p" to appear in your command
without being replaced by the payload, use "%%p". For two percents, use "%%%p",
etc. Hookback removes one of the percents and passes it to the underlying
command.

Waiting
-------
If "wait" is set to true, the Hookback waits for the command to finish running
before responding back to GitHub. This is the default. On success, it will
respond back with "OK", a separate line containing the command that was run, and
finally the output the command (stdout and stderr) on the next line. If the
command does not exit with status 0, a 500 error is returned, the first line of
the response will be "Exit Status #", followed by the output (stdout and stderr)
on the next line.

If "wait" is false, initial work to verify the request will be done, but the
server does not wait for the command to complete before returning "OK" and
the command that was run. It will have no further output. This is useful if
you want to run a long running command. It is up to you to craft a command that
captures output to a file if you wish to preserve it. "sleep 1000" would be a
contrived example of a long running command.
