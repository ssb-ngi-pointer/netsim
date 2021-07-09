package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

type peer struct {
	id      string
	hops    [][]string
	blocked map[string]bool
}

func makePeer(args Args, id string) peer {
	p := peer{id: id}
	p.hops = make([][]string, args.MaxHops+1)
	p.blocked = make(map[string]bool)
	return p
}

func populateHopsAt(args Args, count int, peers map[string]peer) {
	for _, my := range peers {
		for _, friendId := range my.hops[count-1] {
			friend := peers[friendId]
			if len(friend.hops) == 0 || len(friend.hops) < count-1 {
				continue
			}
			for _, hopsFollow := range friend.hops[count-1] {
				// don't add blocked peers to hops
				if _, exists := my.blocked[hopsFollow]; exists && !args.ReplicateBlocked {
					continue
				}
				my.hops[count] = append(my.hops[count], hopsFollow)
			}
		}
	}
}

// TODO: should we include hops[0]? i.e. the peer we are inspecting
func collapse(args Args, peers map[string]peer, blocked map[string]map[string]bool) {
	// prune out duplicates when collapsing the map
	collapsedHops := make(map[string]map[string]bool)
	for id, p := range peers {
		for i := 0; i <= args.MaxHops; i++ {
			collapsedHops[id] = make(map[string]bool)
			for _, otherId := range p.hops[i] {
				collapsedHops[id][otherId] = true
			}
		}
	}

	// massage deduplicated data into a nicer form for later use
	outputMap := make(map[string][]string)
	for id, others := range collapsedHops {
		for otherId := range others {
			// otherId has blocked id -> we should not expect to replicate them
			if blocked[otherId][id] && !args.ReplicateBlocked {
				continue
			}
			outputMap[id] = append(outputMap[id], otherId)
		}
	}

	// persist to disk
	b, err := json.MarshalIndent(outputMap, "", "  ")
	check(err)
	err = os.WriteFile(pathAndFile(args.Outpath, "expectations.json"), b, 0666)
	check(err)
}

func pathAndFile(dirpath, name string) string {
	if strings.HasSuffix(dirpath, "json") {
		dirpath = path.Dir(dirpath)
	}
	return path.Join(dirpath, name)
}

// TO DO:
// * pass in fixturesRoot and use that to derive graphpath
// * ProduceExpectations should output a string? or path to written file?
func ProduceExpectations(args Args, graphpath string) {
	b, err := os.ReadFile(graphpath)
	check(err)

	var v map[string]map[string]interface{}
	err = json.Unmarshal(b, &v)
	check(err)

	// start the party by populating hops 0 via interpreting follow-graph.json:
	// nil => can't deduce info
	// true => peer is followed
	// false => peer is blocked
	peers := make(map[string]peer)
	blocked := make(map[string]map[string]bool)
	for id, relations := range v {
		p := makePeer(args, id)
		blocked[id] = make(map[string]bool)
		p.hops[0] = append(p.hops[0], id)
		for relationId, status := range relations {
			if followed, ok := status.(bool); ok {
				// non-nil relations are followed if status is true
				if followed {
					p.hops[1] = append(p.hops[1], relationId)
					// and blocked if false
				} else {
					p.blocked[relationId] = true
					blocked[id][relationId] = true
				}
			}
		}
		peers[id] = p
	}

	if args.MaxHops >= 2 {
		for i := 2; i <= args.MaxHops; i++ {
			populateHopsAt(args, i, peers)
		}
	}
	collapse(args, peers, blocked)
}

type Args struct {
	MaxHops          int
	ReplicateBlocked bool
	Outpath          string
}

func main() {
	var args Args
	flag.IntVar(&args.MaxHops, "hops", 2, "the default global hops setting")
	flag.BoolVar(&args.ReplicateBlocked, "replicate-blocked", false, "if flag is present, blocked peers will be replicated")
	flag.StringVar(&args.Outpath, "out", "./expectations.json", "the filename and path where the expectations will be dumped")
	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Println("usage:\n  expectations <flags> <path to fixtures folder>")
		os.Exit(0)
	}

	graphpath := pathAndFile(flag.Args()[0], "follow-graph.json")
	ProduceExpectations(args, graphpath)
}
