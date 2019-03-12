/*

twitchClient is a simple CLI app for testing twitchgo API calls.

*/
package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/ddrboxman/twitchgo"
	"github.com/k0kubun/pp"

	"github.com/motemen/go-loghttp"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true

	var clientID string
	var verbosity int

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:        "verbose",
			Value:       0,
			Usage:       "verbosity level - 0, 1, 2",
			Destination: &verbosity,
		},

		cli.StringFlag{
			Name:        "client_id",
			Usage:       "Twitch client ID to authorize requests",
			EnvVar:      "TWITCH_CLIENT_ID",
			Destination: &clientID,
		},
	}

	app.Before = func(c *cli.Context) error {
		if clientID == "" {
			return cli.NewExitError("Twitch client ID was not specified", 1)
		}

		twitchClient = twitch.NewTwitchClient(clientID)

		if verbosity > 0 {
			// Print HTTP requests and possibly bodies
			http.DefaultTransport = &loghttp.Transport{
				Transport: http.DefaultTransport,
				LogRequest: func(req *http.Request) {
					dump, _ := httputil.DumpRequest(req, verbosity > 1)
					fmt.Printf("--> <%s>: \n%s\n", req.URL, string(dump))
				},
				LogResponse: func(resp *http.Response) {
					dump, _ := httputil.DumpResponse(resp, verbosity > 1)
					fmt.Printf("<-- [%d] <%s>:\n%s\n", resp.StatusCode, resp.Request.URL, string(dump))
				},
			}
		}

		return nil
	}

	app.Commands = appCommands

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

var twitchClient twitch.TwitchClient
var appCommands = []cli.Command{
	{
		Name:  "channel",
		Usage: "print a channel's information",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.NewExitError("must supply channel name/ID", 126)
			}

			channel, _ := twitchClient.GetChannel(c.Args().Get(0))

			pp.Println(channel)

			return nil
		},
	},

	{
		Name:  "followers",
		Usage: "print the first page of a channel's followers",
		Action: func(c *cli.Context) error {
			if c.NArg() < 1 {
				return cli.NewExitError("must supply channel ID", 126)
			}

			options := twitch.RequestOptions{
				Limit: 15,
			}

			followers, err := twitchClient.GetFollowersForID(c.Args().Get(0), &options)
			if err != nil {
				return err
			}

			pp.Println(followers)

			return nil
		},
	},
}
