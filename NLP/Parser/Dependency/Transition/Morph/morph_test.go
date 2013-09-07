package Morph

import (
	"chukuparser/Algorithm/Model/Perceptron"
	"chukuparser/NLP/Parser/Dependency"
	"chukuparser/Util"

	G "chukuparser/Algorithm/Graph"
	Transition "chukuparser/Algorithm/Transition"
	T "chukuparser/NLP/Parser/Dependency/Transition"
	NLP "chukuparser/NLP/Types"
	"log"
	"runtime"
	"sort"
	"testing"
)

// sent: HELIM MZHIBIM .
// lattice:
// 0	1	H		_	REL		REL		_								1
// 0	1	H		_	DEF		DEF		_								1
// 0	2	HELIM	_	VB		VB		gen=M|num=S|per=3|tense=PAST	1
// 1	2	ELIM	_	NN		NN		gen=M|num=P						1
// 2	3	MZHIBIM	_	BN		BN		gen=M|num=P|per=A				2
// 2	3	MZHIBIM	_	VB		VB		gen=M|num=P|per=A|tense=BEINONI	2
// 3	4	yyDOT	_	yyDOT	yyDOT	_								3

var (
	TEST_LATTICE NLP.LatticeSentence = NLP.LatticeSentence{
		{"HELIM",
			[]*NLP.EMorpheme{
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{0, 0, 1}, "H", "REL", "REL",
					nil, 1}},
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{1, 0, 1}, "H", "DEF", "DEF",
					nil, 1}},
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{2, 0, 2}, "HELIM", "VB", "VB",
					map[string]string{"gen": "M", "num": "S", "per": "3", "tense": "PAST"}, 1}},
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{3, 1, 2}, "ELIM", "NN", "NN",
					map[string]string{"gen": "M", "num": "P"}, 1}},
			},
			nil,
		},
		{"MZHIBIM",
			[]*NLP.EMorpheme{
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{0, 2, 3}, "MZHIBIM", "BN", "BN",
					map[string]string{"gen": "M", "num": "P", "per": "A"}, 2}},
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{1, 2, 3}, "MZHIBIM", "VB", "VB",
					map[string]string{"gen": "M", "num": "P", "P": "A", "tense": "BEINONI"}, 2}},
			},
			nil,
		},
		{"yyDOT",
			[]*NLP.EMorpheme{
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{0, 3, 4}, "yyDOT", "yyDOT", "yyDOT",
					nil, 3}},
			},
			nil,
		},
	}

	// SENT: HELIM MZHIBIM .
	// gold dep:
	// 0	1	H		_	DEF		DEF		_					2	def		_	_
	// 1	2	ELIM	_	NN		NN		gen=M|num=P			3	subj	_	_
	// 2	3	MZHIBIM	_	BN		BN		gen=M|num=P|per=A	0	prd		_	_
	// 3	4	yyDOT	_	yyDOT	yyDOT	_					3	punct	_	_

	// mapping:
	// 1 HELIM		H:ELIM
	// 2 MZHIBIM	MZHIBIM
	// 3 yyDOT		yyDOT

	TEST_GRAPH *BasicMorphGraph = &BasicMorphGraph{
		T.BasicDepGraph{
			[]NLP.DepNode{
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{0, 0, 0}, "ROOT", "ROOT", "ROOT",
					nil, 0}},
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{1, 0, 1}, "H", "DEF", "DEF",
					nil, 1}},
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{3, 1, 2}, "ELIM", "NN", "NN",
					map[string]string{"gen": "M", "num": "P"}, 1}},
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{0, 2, 3}, "MZHIBIM", "BN", "BN",
					map[string]string{"gen": "M", "num": "P", "per": "A"}, 2}},
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{0, 3, 4}, "yyDOT", "yyDOT", "yyDOT",
					nil, 3}},
			},
			[]*T.BasicDepArc{
				&T.BasicDepArc{Head: 2, RawRelation: "def", Modifier: 1},
				&T.BasicDepArc{Head: 3, RawRelation: "subj", Modifier: 2},
				&T.BasicDepArc{Head: 0, RawRelation: "prd", Modifier: 3},
				&T.BasicDepArc{Head: 3, RawRelation: "punct", Modifier: 4},
			},
		},
		[]*NLP.Mapping{
			// &NLP.Mapping{"ROOT", []*NLP.EMorpheme{}},
			&NLP.Mapping{"HELIM", []*NLP.EMorpheme{
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{1, 0, 1}, "H", "DEF", "DEF",
					nil, 1}},
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{3, 1, 2}, "ELIM", "NN", "NN",
					map[string]string{"gen": "M", "num": "P"}, 1}},
			}},
			&NLP.Mapping{"MZHIBIM", []*NLP.EMorpheme{
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{0, 2, 3}, "MZHIBIM", "BN", "BN",
					map[string]string{"gen": "M", "num": "P", "per": "A"}, 2}},
			}},
			&NLP.Mapping{"yyDOT", []*NLP.EMorpheme{
				&NLP.EMorpheme{Morpheme: NLP.Morpheme{G.BasicDirectedEdge{0, 3, 4}, "yyDOT", "yyDOT", "yyDOT",
					nil, 3}},
			}},
		},
		nil,
	}

	TEST_MORPH_TRANSITIONS []NLP.DepRel = []NLP.DepRel{
		"MD-1", "SH", "LA-def", "SH", "MD-0", "LA-subj", "RA-prd", "MD-0", "RA-punct",
	}

	TEST_RELATIONS []string = []string{
		"advmod", "amod", "appos", "aux",
		"cc", "ccomp", "comp", "complmn",
		"compound", "conj", "cop", "def",
		"dep", "det", "detmod", "gen",
		"ghd", "gobj", "hd", "mod",
		"mwe", "neg", "nn", "null",
		"num", "number", "obj", "parataxis",
		"pcomp", "pobj", "posspmod", "prd",
		"prep", "prepmod", "punct", "qaux",
		"rcmod", "rel", "relcomp", "subj",
		"tmod", "xcomp",
	}

	//ALL RICH FEATURES
	TEST_RICH_FEATURES []string = []string{
		"S0|w|p", "S0|w", "S0|p", "N0|w|p",
		"N0|w", "N0|p", "N1|w|p", "N1|w",
		"N1|p", "N2|w|p", "N2|w", "N2|p",
		"S0|w|p+N0|w|p", "S0|w|p+N0|w",
		"S0|w+N0|w|p", "S0|w|p+N0|p",
		"S0|p+N0|w|p", "S0|w+N0|w",
		"S0|p+N0|p", "N0|p+N1|p",
		"N0|p+N1|p+N2|p", "S0|p+N0|p+N1|p",
		"S0h|p+S0|p+N0|p", "S0|p+S0l|p+N0|p",
		"S0|p+S0r|p+N0|p", "S0|p+N0|p+N0l|p",
		"S0|w|d", "S0|p|d", "N0|w|d", "N0|p|d",
		"S0|w+N0|w|d", "S0|p+N0|p|d",
		"S0|w|vr", "S0|p|vr", "S0|w|vl", "S0|p|vl", "N0|w|vl", "N0|p|vl",
		"S0h|w", "S0h|p", "S0|l", "S0l|w",
		"S0l|p", "S0l|l", "S0r|w", "S0r|p",
		"S0r|l", "N0l|w", "N0l|p", "N0l|l",
		"S0h2|w", "S0h2|p", "S0h|l", "S0l2|w",
		"S0l2|p", "S0l2|l", "S0r2|w", "S0r2|p",
		"S0r2|l", "N0l2|w", "N0l2|p", "N0l2|l",
		"S0|p+S0l|p+S0l2|p", "S0|p+S0r|p+S0r2|p",
		"S0|p+S0h|p+S0h2|p", "N0|p+N0l|p+N0l2|p",
		"S0|w|sr", "S0|p|sr", "S0|w|sl", "S0|p|sl",
		"N0|w|sl", "N0|p|sl"}

	TRANSITIONS_ENUM            *Util.EnumSet
	TEST_MORPH_ENUM_TRANSITIONS []Transition.Transition
	TEST_ENUM_RELATIONS         *Util.EnumSet
	EWord, EPOS, EWPOS          *Util.EnumSet
	SH, RE, LA, RA, MD          Transition.Transition
)

func SetupRelationEnum() {
	if TEST_ENUM_RELATIONS != nil {
		return
	}
	TEST_ENUM_RELATIONS = Util.NewEnumSet(len(TEST_RELATIONS))
	for _, label := range TEST_RELATIONS {
		TEST_ENUM_RELATIONS.Add(label)
	}
}

func SetupSentEnum() {
	EWord, EPOS, EWPOS =
		Util.NewEnumSet(TEST_GRAPH.NumberOfNodes()),
		Util.NewEnumSet(7), // 6 Lattice POS + ROOT
		Util.NewEnumSet(8) // 7 Lattice WPOS + ROOT
	var (
		morph *NLP.EMorpheme
	)
	for _, node := range TEST_GRAPH.Nodes {
		morph = node.(*NLP.EMorpheme)
		morph.EForm, _ = EWord.Add(morph.Form)
		morph.EPOS, _ = EPOS.Add(morph.CPOS)
		morph.EFCPOS, _ = EWPOS.Add([2]string{morph.Form, morph.CPOS})
	}
	for _, arc := range TEST_GRAPH.Arcs {
		arc.Relation, _ = TEST_ENUM_RELATIONS.Add(arc.RawRelation)
	}
	for _, lattice := range TEST_LATTICE {
		for _, morph := range lattice.Morphemes {
			morph.EForm, _ = EWord.Add(morph.Form)
			morph.EPOS, _ = EPOS.Add(morph.CPOS)
			morph.EFCPOS, _ = EWPOS.Add([2]string{morph.Form, morph.CPOS})
		}
	}
}

const APPROX_MORPH_TRANSITIONS = 30

func SetupMorphTransEnum() {
	TRANSITIONS_ENUM = Util.NewEnumSet(len(TEST_RELATIONS)*2 + 2 + APPROX_MORPH_TRANSITIONS)
	iSH, _ := TRANSITIONS_ENUM.Add(NLP.DepRel("SH"))
	iRE, _ := TRANSITIONS_ENUM.Add(NLP.DepRel("RE"))
	SH = Transition.Transition(iSH)
	RE = Transition.Transition(iRE)
	LA = RE + 1
	for _, transition := range TEST_RELATIONS {
		TRANSITIONS_ENUM.Add(NLP.DepRel("LA-" + transition))
	}
	RA = Transition.Transition(TRANSITIONS_ENUM.Len())
	for _, transition := range TEST_RELATIONS {
		TRANSITIONS_ENUM.Add(NLP.DepRel("RA-" + transition))
	}
	MD = Transition.Transition(TRANSITIONS_ENUM.Len())
	TEST_MORPH_ENUM_TRANSITIONS = make([]Transition.Transition, len(TEST_MORPH_TRANSITIONS))
	for i, transition := range TEST_MORPH_TRANSITIONS {
		index, _ := TRANSITIONS_ENUM.Add(transition)
		TEST_MORPH_ENUM_TRANSITIONS[i] = Transition.Transition(index)
	}
}

func SetupTestEnum() {
	SetupRelationEnum()
	SetupSentEnum()
	SetupMorphTransEnum()
}

func TestOracle(t *testing.T) {
	SetupTestEnum()
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("Testing Oracle")
	mconf := new(MorphConfiguration)
	mconf.Init(TEST_LATTICE)
	conf := Transition.Configuration(mconf)
	arcmorph := &ArcEagerMorph{
		ArcEager: T.ArcEager{
			ArcStandard: T.ArcStandard{
				SHIFT:       SH,
				LEFT:        LA,
				RIGHT:       RA,
				Relations:   TEST_ENUM_RELATIONS,
				Transitions: TRANSITIONS_ENUM,
			},
			REDUCE: RE},
		MD: MD,
	}
	arcmorph.AddDefaultOracle()
	trans := Transition.TransitionSystem(arcmorph)
	trans.Oracle().SetGold(TEST_GRAPH)

	goldTrans := TEST_MORPH_ENUM_TRANSITIONS
	for !conf.Terminal() {
		oracle := trans.Oracle()
		transition := oracle.Transition(conf)
		log.Println("Chose transition, gold:", transition, goldTrans)
		if transition != goldTrans[0] {
			t.Error("Gold is:", goldTrans[0], "got", transition)
			return
		}
		conf = trans.Transition(conf, transition)
		goldTrans = goldTrans[1:]
	}
	log.Println("Done testing Oracle")
	// log.Println("\n", conf.GetSequence().String())
}

func TestDeterministic(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("Testing Deterministic")
	runtime.GOMAXPROCS(runtime.NumCPU())
	extractor := new(T.GenericExtractor)
	// verify load
	for _, feature := range TEST_RICH_FEATURES {
		if err := extractor.LoadFeature(feature); err != nil {
			t.Error("Failed to load feature", err.Error())
			t.FailNow()
		}
	}
	arcSystem := &ArcEagerMorph{
		ArcEager: T.ArcEager{
			ArcStandard: T.ArcStandard{
				SHIFT:       SH,
				LEFT:        LA,
				RIGHT:       RA,
				Relations:   TEST_ENUM_RELATIONS,
				Transitions: TRANSITIONS_ENUM,
			},
			REDUCE: RE},
		MD: MD,
	}

	conf := &MorphConfiguration{SimpleConfiguration: T.SimpleConfiguration{
		EWord:  EWord,
		EPOS:   EPOS,
		EWPOS:  EWPOS,
		ERel:   TEST_ENUM_RELATIONS,
		ETrans: TRANSITIONS_ENUM,
	},
	}

	arcSystem.AddDefaultOracle()
	transitionSystem := Transition.TransitionSystem(arcSystem)
	deterministic := &T.Deterministic{
		TransFunc:          transitionSystem,
		FeatExtractor:      extractor,
		ReturnModelValue:   true,
		ReturnSequence:     true,
		ShowConsiderations: false,
		Base:               conf,
		NoRecover:          true,
	}

	decoder := Perceptron.EarlyUpdateInstanceDecoder(deterministic)
	updater := new(Perceptron.AveragedStrategy)

	goldInstances := []Perceptron.DecodedInstance{
		&Perceptron.Decoded{Perceptron.Instance(TEST_LATTICE), TEST_GRAPH}}

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	perceptron.Init()
	goldModel := Dependency.ParameterModel(&T.PerceptronModel{perceptron})
	graph := TEST_GRAPH
	graph.Lattice = TEST_LATTICE

	_, goldParams := deterministic.ParseOracle(graph, nil, goldModel)
	goldSequence := goldParams.(*T.ParseResultParameters).Sequence

	// train with increasing iterations
	convergenceIterations := []int{1, 2, 8, 16, 32}
	convergenceSharedSequence := make([]int, 0, len(convergenceIterations))
	for _, iterations := range convergenceIterations {
		perceptron.Iterations = iterations
		// perceptron.Log = true
		perceptron.Init()

		// deterministic.ShowConsiderations = true
		perceptron.Train(goldInstances)

		model := Dependency.ParameterModel(&T.PerceptronModel{perceptron})
		// deterministic.ShowConsiderations = true
		_, params := deterministic.Parse(TEST_LATTICE, nil, model)
		seq := params.(*T.ParseResultParameters).Sequence
		// log.Println(seq)
		sharedSteps := goldSequence.SharedTransitions(seq)
		convergenceSharedSequence = append(convergenceSharedSequence, sharedSteps)
	}

	// verify convergence
	if !sort.IntsAreSorted(convergenceSharedSequence) || convergenceSharedSequence[0] == convergenceSharedSequence[len(convergenceSharedSequence)-1] {
		t.Error("Model not converging, shared sequences lengths:", convergenceSharedSequence)
	}
	log.Println("Done Testing Deterministic")
}

func TestSimpleBeam(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)
	log.Println("Testing Simple Beam")
	runtime.GOMAXPROCS(runtime.NumCPU())
	// runtime.GOMAXPROCS(1)
	extractor := new(T.GenericExtractor)
	// verify load
	for _, feature := range TEST_RICH_FEATURES {
		if err := extractor.LoadFeature(feature); err != nil {
			t.Error("Failed to load feature", err.Error())
			t.FailNow()
		}
	}
	arcSystem := &ArcEagerMorph{
		ArcEager: T.ArcEager{
			ArcStandard: T.ArcStandard{
				SHIFT:       SH,
				LEFT:        LA,
				RIGHT:       RA,
				Relations:   TEST_ENUM_RELATIONS,
				Transitions: TRANSITIONS_ENUM,
			},
			REDUCE: RE},
		MD: MD,
	}
	arcSystem.AddDefaultOracle()
	transitionSystem := Transition.TransitionSystem(arcSystem)

	conf := &MorphConfiguration{SimpleConfiguration: T.SimpleConfiguration{
		EWord:  EWord,
		EPOS:   EPOS,
		EWPOS:  EWPOS,
		ERel:   TEST_ENUM_RELATIONS,
		ETrans: TRANSITIONS_ENUM,
	},
	}

	beam := &T.Beam{
		TransFunc:     transitionSystem,
		FeatExtractor: extractor,
		Base:          conf,
		NumRelations:  arcSystem.Relations.Len(),
	}

	decoder := Perceptron.EarlyUpdateInstanceDecoder(beam)
	updater := new(Perceptron.AveragedStrategy)

	perceptron := &Perceptron.LinearPerceptron{Decoder: decoder, Updater: updater}
	perceptron.Init()
	graph := TEST_GRAPH
	graph.Lattice = TEST_LATTICE

	// get gold parse
	goldModel := Dependency.ParameterModel(&T.PerceptronModel{perceptron})
	deterministic := &T.Deterministic{
		TransFunc:          transitionSystem,
		FeatExtractor:      extractor,
		ReturnModelValue:   true,
		ReturnSequence:     true,
		ShowConsiderations: false,
		Base:               conf,
		NoRecover:          true,
	}
	_, goldParams := deterministic.ParseOracle(graph, nil, goldModel)
	goldSequence := goldParams.(*T.ParseResultParameters).Sequence

	goldInstances := []Perceptron.DecodedInstance{
		&Perceptron.Decoded{Perceptron.Instance(TEST_LATTICE), goldSequence[0]}}

	// train with increasing iterations
	beam.ConcurrentExec = true
	beam.ReturnSequence = true

	convergenceIterations := []int{1, 4, 16}
	beamSizes := []int{1, 4, 16, 64}
	for _, beamSize := range beamSizes {
		beam.Size = beamSize

		convergenceSharedSequence := make([]int, 0, len(convergenceIterations))
		for _, iterations := range convergenceIterations {
			perceptron.Iterations = iterations
			// perceptron.Log = true
			perceptron.Init()

			// deterministic.ShowConsiderations = true
			perceptron.Train(goldInstances)

			model := Dependency.ParameterModel(&T.PerceptronModel{perceptron})
			beam.ReturnModelValue = false
			_, params := beam.Parse(TEST_LATTICE, nil, model)
			sharedSteps := 0
			if params != nil {
				seq := params.(*T.ParseResultParameters).Sequence
				sharedSteps = goldSequence.SharedTransitions(seq)
			}
			convergenceSharedSequence = append(convergenceSharedSequence, sharedSteps)
		}
		if len(convergenceSharedSequence) != len(convergenceIterations) {
			t.Error("Not enough examples in shared sequence samples")
		}
		// verify convergence
		log.Println("Shared Sequence For Beam", beamSize, convergenceSharedSequence)
		if !sort.IntsAreSorted(convergenceSharedSequence) || convergenceSharedSequence[len(convergenceSharedSequence)-1] == 0 {
			t.Error("Model not converging, shared sequences lengths:", convergenceSharedSequence)
		}
	}
	log.Println("Done Testing Simple Beam")
}
