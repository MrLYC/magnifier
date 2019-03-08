package magnifier

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/emirpasic/gods/trees/binaryheap"
	"github.com/google/subcommands"
	"github.com/mrlyc/magnifier/binding"
	"github.com/mrlyc/magnifier/logging"
	"github.com/mrlyc/magnifier/sego"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Command :
type Command struct {
	dictionary   string
	flag         string
	minFrequency int
	count        int
	segmenter    *sego.Segmenter
	flagSet      *hashset.Set
}

// Name :
func (*Command) Name() string {
	return "analysis"
}

// Synopsis :
func (*Command) Synopsis() string {
	return "analysis documents"
}

// Usage :
func (*Command) Usage() string {
	return `analysis
  analysis documents.
`
}

type word struct {
	flag    string
	text    string
	counter int
}

func (c *Command) initSegmenter(dictionaries ...string) {
	logger := logging.GetLogger()

	readers := make([]io.Reader, 0, len(dictionaries))
	readClosers := make([]io.ReadCloser, 0, len(dictionaries))
	for _, path := range dictionaries {
		fi, err := os.Stat(path)
		if !os.IsNotExist(err) && !fi.IsDir() {
			file, err := os.Open(path)
			if err == nil {
				logger.Infof("dictionary file found: %v", path)
				readClosers = append(readClosers, file)
				readers = append(readers, file)
				continue
			}
		}

		data, err := binding.Asset(path)
		if err != nil {
			logger.Fatalf("load dictionary %v failed", path)
		}

		logger.Infof("built in dictionary: %v", path)
		reader := ioutil.NopCloser(bytes.NewReader(data))

		readClosers = append(readClosers, reader)
		readers = append(readers, reader)
	}

	c.segmenter.Load(readers...)

	for _, reader := range readClosers {
		err := reader.Close()
		if err != nil {
			logger.Errorf("close dictionary %v error %v", reader, err)
		}
	}
}

// SetFlags :
func (c *Command) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.dictionary, "dictionary", "data/dictionary.txt", "dictionary path")
	f.StringVar(&c.flag, "flag", "nz,ng,nt,nr,nrfg,n,ns,nrt", "available flag(tg,l,ng,t,mq,e,ag,an,h,ug,ns,n,c,uz,vg,r,vq,f,p,dg,g,k,vd,m,zg,uv,s,z,nrt,uj,rz,ud,nt,rg,nrfg,i,df,ad,v,q,vi,rr,nz,o,mg,u,j,ul,nr,a,vn,b,d,y)")
	f.IntVar(&c.minFrequency, "min-frequency", 0, "min frequency")
	f.IntVar(&c.count, "count", 5, "word count")
}

func (c *Command) analysisFiles(path ...string) {
	logger := logging.GetLogger()

	counter := make(map[string]*word)

	for _, p := range path {
		data, err := ioutil.ReadFile(p)
		if err != nil {
			logger.Errorf("read file %v error %v", path, err)
			return
		}

		for _, s := range c.segmenter.Segment(data) {
			token := s.Token()
			pos := token.Pos()
			if !c.flagSet.Contains(pos) {
				continue
			}
			if token.Frequency() < c.minFrequency {
				continue
			}
			text := token.Text()
			w, ok := counter[text]
			if ok {
				w.counter ++
			} else {
				counter[text] = &word{
					text:    text,
					counter: 1,
					flag:    pos,
				}
			}
		}
	}

	heap := binaryheap.NewWith(func(a, b interface{}) int {
		aAsserted := a.(*word)
		bAsserted := b.(*word)
		switch {
		case aAsserted.counter > bAsserted.counter:
			return -1
		case aAsserted.counter < bAsserted.counter:
			return 1
		default:
			return 0
		}
	})
	for _, w := range counter {
		heap.Push(w)
	}
	for i := 0; i < c.count; i++ {
		v, ok := heap.Pop()
		if !ok {
			break
		}
		w := v.(*word)
		fmt.Printf("%s/%s ", w.text, w.flag)
	}
	fmt.Printf("\n")
}

func (c *Command) completer(d prompt.Document) []prompt.Suggest {
	s := make([]prompt.Suggest, 0)
	files, err := filepath.Glob(fmt.Sprintf("%s*", d.CurrentLine()))
	if err == nil {
		for _, file := range files {
			s = append(s, prompt.Suggest{
				Text: file,
			})
		}
	}
	return s
}

// Execute :
func (c *Command) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	c.segmenter = new(sego.Segmenter)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		c.initSegmenter(strings.Split(c.dictionary, ",")...)
		wg.Done()
	}()

	c.flagSet = hashset.New()
	for _, f := range strings.Split(c.flag, ",") {
		c.flagSet.Add(f)
	}

	for {
		line := prompt.Input("", c.completer)
		if line == "" {
			break
		}

		files := make([]string, 0)
		for _, f := range strings.Split(line, ",") {
			files = append(files, f)
		}

		wg.Wait()

		c.analysisFiles(files...)
	}

	return subcommands.ExitSuccess
}
