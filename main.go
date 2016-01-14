package main

import (
	"log"
	"os"
	"strconv"

	"github.com/codegangsta/cli"
	"github.com/therealbill/libredis/client"
	"github.com/therealbill/libredis/structures"
)

var (
	app           *cli.App
	sentinel_list cli.StringSlice
)

func main() {
	app = cli.NewApp()
	app.Name = "correct-pod-configs"
	app.Usage = "Correct all pods' down-after-millisecond value in a given constellation"
	app.Version = "1.1"
	app.EnableBashCompletion = true
	author := cli.Author{Name: "Bill Anderson", Email: "bill.anderson@rackspace.com"}
	app.Authors = append(app.Authors, author)
	sentinel_list = cli.StringSlice{}
	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name:  "sentinels,s",
			Usage: "Addresses of the sentinel nodes",
			Value: &sentinel_list,
		},
		cli.StringFlag{
			Name:   "targetvalue,t",
			Usage:  "The value to set down-after-milliseconds to",
			Value:  "3000",
			EnvVar: "CD_TARGET",
		},
		cli.IntFlag{
			Name:   "ignore,i",
			Usage:  "Skip over pods with this s their down-after-milliseconds value",
			Value:  345600000,
			EnvVar: "CD_IGNORE",
		},
		cli.BoolFlag{
			Name:  "commit,c",
			Usage: "If enabled, will actually commit the change in Sentinel. Default is off for checking.",
		},
		cli.BoolFlag{
			Name:  "all,a",
			Usage: "If enabled with -c will the change to ALL pods in the constellation",
		},
		cli.BoolFlag{
			Name:  "verbose,d",
			Usage: "Be a bit more chatty",
		},
	}
	app.Action = setTimeout
	app.Run(os.Args)

}

func setTimeout(c *cli.Context) {
	sentinels := c.StringSlice("sentinels")
	if len(sentinels) == 0 {
		log.Fatal("Need a list of Sentinel addresses")
	}
	target := c.String("targetvalue")
	if target == "" {
		log.Fatal("Need a value to set them to there, Hoss.")
	}
	for _, s := range sentinels {
		spods, conn, err := getPodListFromSentinel(s)
		if err != nil {
			log.Printf("Skipping sentinel %s", s)
			continue
		}
		defaults := 0
		corrects := 0
		others := 0
		ignores := 0
		itarget, _ := strconv.Atoi(target)
		for _, pod := range spods {
			if c.Bool("all") {
				if pod.DownAfterMilliseconds == itarget {
					if c.Bool("verbose") {
						log.Printf("[%s] Pod %s is correct, skipping.", s, pod.Name)
					}
					corrects++
				} else {
					if c.Bool("verbose") {
						log.Printf("[%s] Pod %s is incorrect.", s, pod.Name)
					}
					others++
					if c.Bool("commit") {
						log.Printf("[%s] Updating pod %s to %s from %d", s, pod.Name, target, pod.DownAfterMilliseconds)
						_ = conn.SentinelSetString(pod.Name, "down-after-milliseconds", target)
					}
				}
			} else {
				switch pod.DownAfterMilliseconds {
				case itarget:
					corrects++
					if c.Bool("verbose") {
						log.Printf("[%s] Pod %s is correct", s, pod.Name)
					}
				case 30000:
					defaults++
					if c.Bool("verbose") {
						log.Printf("[%s] Pod %s has Redis default", s, pod.Name)
					}
					if c.Bool("commit") {
						log.Printf("[%s] Updating pod %s to %s from default", s, pod.Name, target)
						_ = conn.SentinelSetString(pod.Name, "down-after-milliseconds", target)
					}
				case c.Int("ignore"):
					ignores++
					if c.Bool("verbose") {
						log.Printf("[%s] Pod %s has ignores value", s, pod.Name)
					}
				default:
					others++
					log.Printf("[%s] Pod %s has invalid value %d", s, pod.Name, pod.DownAfterMilliseconds)
				}
			}
		}
		log.Printf("[%s] Correct: %d Default: %d Resizes: %d Other: %d", s, corrects, defaults, ignores, others)
	}

}

func getPodListFromSentinel(ip string) (pods []structures.MasterInfo, conn *client.Redis, err error) {
	conn, err = client.Dial(ip, 26379)
	if err != nil {
		log.Printf("Unable to connect to %s. Error='%v'", ip, err)
		return
	}
	pods, err = conn.SentinelMasters()
	return
}
