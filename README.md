A simple GitHub WebHook server.

Provide a configuration file. Tell it what command to run for what events and on
which repositories. It will do just that. That's it.

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
          "cmd": ["echo", "pinged"]
        }
      }
    }
  }
}

If you're familiar with GitHub's webhooks, the breakdown is fairly self
explanatory. After enabling webhooks on the "mankyd/hookback" repository,
this server will run "echo pinged" anytime it receives a "ping" event from
GitHub. It expects the webhook secret to be set to "asdf" and will verify this,
rejecting any requests that are not properly signed with this secret.

The server section has only two options, both shown. By default, the server
listens on all interfaces ("0.0.0.0") and port "3015". If this is acceptable,
you are free to remove the "server" section from the config entirely.

If "wait" is set to true, the Hookback waits for the command to finish running
before responding back to GitHub. This is the default. On success, it will
respond back with "OK", a separate line containing the command that was run, and
finally the output the command (stdout and stderr) on the next line. If the
command does not exit with status 0, a 500 error is returned, the first line of
the response will be "Exit Status #", followed by the output (stdout and stderr)
on the next line.

If "wait" is false, initial work to verify the request will be done, but the 
server does not wait for the command to complete before returning "OK", followed
by the command that was run. It will have no further output. This is useful if 
you want to run a long running command. It is up to you to craft a command that
captures output if you wish to preserve it. "sleep 1000" would be a contrived
example of a long running command.
