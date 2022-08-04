package environment

import (
	"github.com/gojinja/gojinja/src/nodes"
	"github.com/gojinja/gojinja/src/utils/set"
)

type symbols struct {
	level  int
	parent *symbols
	refs   map[string]string
	loads  map[string]any
	stores set.Set[string]
}

func newSymbols(parent *symbols, level *int) *symbols {
	var lvl int
	if level != nil {
		lvl = *level
	} else {
		if parent == nil {
			lvl = 0
		} else {
			lvl = parent.level + 1
		}
	}

	return &symbols{
		level:  lvl,
		parent: parent,
		refs:   make(map[string]string),
		loads:  make(map[string]any),
		stores: set.New[string](),
	}
}

func (s *symbols) analyzeNode(node nodes.Node) {
	// TODO port from jinja
}

type frame struct {
	evalCtx            *evalContext
	parent             *frame
	symbols            *symbols
	requireOutputCheck bool
	buffer             *string
	block              *string // the name of the block we're in, otherwise nil
	toplevel           bool    // a toplevel frame is the root + soft frames such as if conditions.
	// the root frame is basically just the outermost frame, so no if
	// conditions. This information is used to optimize inheritance situations.
	rootlevel bool
	// variables set inside of loops and blocks should not affect outer frames,
	// but they still needs to be kept track of as part of the active context.
	loopFrame  bool
	blockFrame bool
	// track whether the frame is being used in an if-statement or conditional
	// expression as it determines which errors should be raised during runtime
	// or compile time.
	softFrame bool
}

func newFrame(evalCtx *evalContext, parent *frame, level *int) *frame {
	f := &frame{
		evalCtx: evalCtx,
		parent:  parent,
	}

	if parent == nil {
		f.symbols = newSymbols(nil, level)
	} else {
		f.symbols = newSymbols(parent.symbols, level)
		f.requireOutputCheck = parent.requireOutputCheck
		f.buffer = parent.buffer
		f.block = parent.block
	}

	return f
}
