package gomail

import (
	"crypto/tls"
	"log"

	"github.com/urfave/cli/v2"
	"gopkg.in/gomail.v2"
)

var encryption = "OMZNGORYYAXGLTUP"

type M_cli struct {
	From        string
	To          []string
	Subject     string
	Body        string
	Enclosure   string
	Smtp_domain string
	Port        int
	User        string
	Password    string
}

func (c *M_cli) SendMail() error {
	m := gomail.NewMessage()
	// m.SetHeader("From", "kongbaiai2@126.com")
	// m.SetHeader("To", "wull@yipeng888.com", "kongbaiai2@126.com")
	// // m.SetAddressHeader("Cc", "wull@yipeng888.com", "Dan")
	// m.SetHeader("Subject", "Hello!")
	// m.SetBody("text/html", "Hello <b>Bob</b> and <i>Cora</i>!")
	// // m.Attach("/home/Alex/lolcat.jpg")

	// d := gomail.NewDialer("smtp.126.com", 465, "kongbaiai2@126.com", "OMZNGORYYAXGLTUP")
	m.SetHeader("From", c.From)
	m.SetHeader("To", c.To...)
	m.SetHeader("Subject", c.Subject)
	m.SetBody("text/html", c.Body)
	if c.Enclosure != "" {
		m.Attach(c.Enclosure)
	}
	// log.Printf("%v", m.GetHeader("To"))
	d := gomail.NewDialer(c.Smtp_domain, c.Port, c.User, c.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Send the email failed: %v", err)
		return err
		// panic(err)
	}
	return nil
}
func GoMail(c *cli.Context, sliceFlag *cli.StringSlice) error {
	// if c.String("password") == "" {
	// 	return nil
	// }
	to := []string{}
	if len(c.StringSlice("to")) == 0 {
		sliceFlag.Set("403853934@qq.com")
		to = sliceFlag.Value()
	} else {
		to = c.StringSlice("to")
	}

	m_cli := &M_cli{
		From:        c.String("from"),
		To:          to,
		Subject:     c.String("subject"),
		Body:        c.String("body"),
		Enclosure:   c.String("enclosure"),
		Smtp_domain: c.String("smtp"),
		Port:        c.Int("port"),
		User:        c.String("user"),
		Password:    c.String("password"),
	}

	log.Printf("send to:%v", to)
	if m_cli.Password == "" && c.String("key") == "yilang" {
		m_cli.Password = "OMZNGORYYAXGLTUP"
	}

	err := m_cli.SendMail()
	if err != nil {
		return err
	}

	return nil
}

func AddSendMail(goCom []*cli.Command) []*cli.Command {
	sliceFlag := &cli.StringSlice{} //&[]string{"kongbaiai2@qq.com"}
	Command := &cli.Command{
		Name: "mail",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "from",
				Aliases: []string{"f"},
				Usage:   "Sender mail name case: 123@126.com",
				Value:   "kongbaiai2@126.com",
			},
			// &cli.StringFlag{
			&cli.StringSliceFlag{
				Name:    "to",
				Aliases: []string{"t"},
				Usage:   "requisite recipient mail name case: 123@126.com",
				Value:   sliceFlag,
			},
			&cli.StringFlag{
				Name:    "subject",
				Aliases: []string{"su"},
				Usage:   "Subject mail head case: hello",
				Value:   "zabbix",
			},
			&cli.StringFlag{
				Name:    "body",
				Aliases: []string{"b"},
				Usage:   "Subject mail body case: text",
				Value:   "describe",
			},
			&cli.StringFlag{
				Name:    "smtp",
				Aliases: []string{"s"},
				Usage:   "specify mail server domain: smtp.126.com",
				Value:   "smtp.126.com",
			},
			&cli.IntFlag{
				Name:    "port",
				Aliases: []string{"po"},
				Usage:   "specify mail server domain: 465",
				Value:   465,
			},
			&cli.StringFlag{
				Name:    "user",
				Aliases: []string{"u"},
				Usage:   "specify mail user: k@126.com",
				Value:   "kongbaiai2@126.com",
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"p"},
				Usage:   "specify mail password: encryption",
				// Value:   "encryption",
			},
			&cli.StringFlag{
				Name:    "key",
				Aliases: []string{"k"},
				Usage:   "specify mail key: encryption",
			},
			&cli.StringFlag{
				Name:    "enclosure",
				Aliases: []string{"e"},
				Usage:   "specify mail enclosure: ./1.jpg",
			},
		},
		Action: func(c *cli.Context) error {

			GoMail(c, sliceFlag)
			return nil
		},
	}

	goCom = append(goCom, Command)
	return goCom
}
