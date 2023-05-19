package registration

import (
	"log"

	"github.com/urfave/cli/v2"
)

var matchers = make(map[string]Commander)

type Acli struct {
	app *cli.App
}

type Commander interface {
	Appends(gCommands []*cli.Command) []*cli.Command
}

func NewApp() *Acli {
	return &Acli{
		app: cli.NewApp(),
	}
}

func (a *Acli) AppInfo() *Acli {
	a.app.Name = "gitapply"
	a.app.Usage = "custom tools"
	// a.app.UsageText= ""
	return a
}

func (a *Acli) Run(Args []string) error {
	return a.app.Run(Args)
}

func Run(args []string) {
	app := NewApp()
	app.AppInfo()

	for _, command := range matchers {
		app.app.Commands = command.Appends(app.app.Commands)
	}
	app.Run(args)
}

func Register(feedType string, matcher Commander) {
	if _, exists := matchers[feedType]; exists {
		log.Fatal(feedType, "command already registered")
	}
	log.Println("Register", feedType, "command")
	matchers[feedType] = matcher
}
