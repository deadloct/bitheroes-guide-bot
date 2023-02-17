package cmd

type GuideOption string

const (
	AuthorOption    GuideOption = "author"
	FamOption       GuideOption = "fam"
	GuideTypeOption GuideOption = "guide-type"
	SetOption       GuideOption = "set"

	GuidePrefix = "guides-"
)

var (
	GuideOptions = []GuideOption{AuthorOption, FamOption, GuideTypeOption, SetOption}
)

type Guides struct {
	GuideIndex map[GuideOption]map[string][]*JSONCommandOption

	Authors    []string
	Familiars  []string
	Sets       []string
	GuideTypes []string
}

func NewGuides(cmds []JSONCommand) *Guides {
	g := &Guides{
		GuideIndex: make(map[GuideOption]map[string][]*JSONCommandOption),
	}

	for _, opt := range GuideOptions {
		g.GuideIndex[opt] = make(map[string][]*JSONCommandOption)
	}

	for _, cmd := range cmds {
		for _, guide := range cmd.Guides {
			for _, set := range guide.Sets {

			}
		}
	}
}
