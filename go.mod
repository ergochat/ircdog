module github.com/ergochat/ircdog

go 1.19

require (
	github.com/chzyer/readline v1.5.1
	github.com/docopt/docopt-go v0.0.0-20160216232012-784ddc588536
	github.com/ergochat/irc-go v0.3.0
	github.com/gorilla/websocket v1.5.0
	github.com/jwalton/go-supportscolor v1.1.0
)

require (
	golang.org/x/sys v0.0.0-20220310020820-b874c991c1a5 // indirect
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d // indirect
)

replace github.com/chzyer/readline => github.com/slingamn/readline v0.0.0-20230213051602-7bb0e056741f
