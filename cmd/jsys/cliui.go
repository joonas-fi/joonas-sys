package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type Void struct{}

type process struct {
	phases         []*Step
	activePhase    *Step
	activePhaseIdx int
	previousPhase  *Step
}

func newProcess(steps []*Step) *process {
	p := &process{
		phases: steps,
	}

	p.activePhase = p.phases[p.activePhaseIdx]

	return p
}

func (p *process) advanceToNextStep() bool {
	p.activePhaseIdx++

	if p.activePhaseIdx >= len(p.phases) {
		return false
	}

	p.previousPhase = p.activePhase
	p.activePhase = p.phases[p.activePhaseIdx]

	return true
}

func (p *Step) logTail(n int) []string {
	startIdx := func() int {
		if len(p.logLines) <= n {
			return 0
		} else {
			return len(p.logLines) - n
		}
	}()

	return p.logLines[startIdx:len(p.logLines)]
}

func displayFancyUI(ctx context.Context, nextStep chan Void, steps []*Step, appendLogLine chan string) error {
	s, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := s.Init(); err != nil {
		return err
	}
	defer s.Fini()

	proc := newProcess(steps)

	styleNormal := tcell.StyleDefault
	styleBold := tcell.StyleDefault.Bold(true)

	// spinner:=newSpinner(spinnerAnimToggle)
	spinner := newSpinner(spinnerAnimDots12)

	drawScreen := func() {
		width, _ := s.Size()
		s.Clear()

		yPos := 0

		x := 0
		for i, phase := range proc.phases[max(proc.activePhaseIdx-1, 0):] {
			if phase == proc.activePhase {
				x = emitStr(s, x, yPos, styleBold, phase.FriendlyName)
			} else {
				x = emitStr(s, x, yPos, styleNormal, phase.FriendlyName)
			}

			if i != len(proc.phases)-1 {
				x = emitStr(s, x, yPos, tcell.StyleDefault, " > ")
			}
		}
		yPos += 1

		pbar := ProgressBar(int(float64(proc.activePhaseIdx)/float64((len(proc.phases)-1))*100), width, ProgressBarDefaultTheme())
		emitStr(s, 0, yPos, tcell.StyleDefault, pbar)

		yPos += 1

		boxHeight := 6
		boxHeightInclBorders := func() int {
			return boxHeight + 2
		}

		if proc.previousPhase != nil {
			box(s, 0, yPos, width, boxHeightInclBorders(), fmt.Sprintf("Previous (%s)", proc.previousPhase.FriendlyName))
			for idx, line := range proc.previousPhase.logTail(boxHeight) {
				emitStr(s, 1, yPos+1+idx, tcell.StyleDefault, line[0:min(len(line), width-2)])
			}
		}

		yPos += boxHeightInclBorders()

		yPos += 1 // empty line

		boxHeight = 20

		activeBoxTitle := fmt.Sprintf("%s %s ", proc.activePhase.FriendlyName, spinner.Get())

		box(s, 0, yPos, width, boxHeightInclBorders(), activeBoxTitle)
		for idx, line := range proc.activePhase.logTail(boxHeight) {
			emitStr(s, 1, yPos+1+idx, tcell.StyleDefault, line[0:min(len(line), width-2)])
		}

		s.Show()
	}

	drawScreen()

	screenEvent := make(chan tcell.Event, 1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			screenEvent <- s.PollEvent()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case line := <-appendLogLine:
			step := proc.activePhase

			step.logLines = append(step.logLines, line)

			drawScreen()
		case <-nextStep:
			proc.advanceToNextStep()

			drawScreen()
		case <-spinner.needsUpdate.C:
			drawScreen()
		case evGeneric := <-screenEvent:
			switch ev := evGeneric.(type) {
			case *tcell.EventResize:
				s.Sync()

				drawScreen()
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyEscape {
					s.Fini()
					return nil
				}
			}
		}
	}
}

func emitStr(s tcell.Screen, x, y int, style tcell.Style, str string) int {
	for _, c := range str {
		var comb []rune
		w := runewidth.RuneWidth(c)
		if w == 0 {
			comb = []rune{c}
			c = ' '
			w = 1
		}
		s.SetContent(x, y, c, comb, style)
		x += w
	}

	return x
}

func horizline(s tcell.Screen, x, y, width int) {
	for i := 0; i < width; i++ {
		s.SetContent(x+i, y, '─', nil, tcell.StyleDefault)
	}
}

func vertline(s tcell.Screen, x, y, height int) {
	for i := 0; i < height; i++ {
		s.SetContent(x, y+i, '│', nil, tcell.StyleDefault)
	}
}

func box(s tcell.Screen, x, y, width, height int, legend string) {
	horizline(s, x, y, width)
	horizline(s, x, y+height-1, width)
	vertline(s, x, y, height)
	vertline(s, x+width-1, y, height)

	// corners
	s.SetContent(x, y, '┌', nil, tcell.StyleDefault)
	s.SetContent(x+width-1, y, '┐', nil, tcell.StyleDefault)
	s.SetContent(x, y+height-1, '└', nil, tcell.StyleDefault)
	s.SetContent(x+width-1, y+height-1, '┘', nil, tcell.StyleDefault)

	if legend != "" {
		emitStr(s, x+2, y, tcell.StyleDefault, legend)
	}
}

type AnimFrameGetter interface {
	GetFrames() []string
}

type animFrameLength1 string

var _ AnimFrameGetter = (*animFrameLength1)(nil)

func (s animFrameLength1) GetFrames() []string {
	frames := []string{}
	for _, rune := range []rune(s) { // needs to be runes, because not all symbols are 1-byte long
		frames = append(frames, string(rune))
	}

	return frames
}

type animFrameLength2 string

var _ AnimFrameGetter = (*animFrameLength2)(nil)

func (a animFrameLength2) GetFrames() []string {
	frames := []string{}
	asRunes := []rune(a)
	for i := 0; i < len(asRunes); i += 2 {
		frames = append(frames, string(asRunes[i:i+2]))
	}

	return frames
}

// https://raw.githubusercontent.com/sindresorhus/cli-spinners/master/spinners.json
const (
	spinnerAnimSpinningLine animFrameLength1 = `-\|/`
	// thanks https://stackoverflow.com/a/2685827
	spinnerAnimBraille            animFrameLength1 = "⣾⣽⣻⢿⡿⣟⣯⣷⠁⠂⠄⡀⢀⠠⠐⠈"
	spinnerAnimPie                animFrameLength1 = "◴◷◶◵"
	spinnerAnimSpinningRectangle1 animFrameLength1 = "◰◳◲◱"
	spinnerAnimSpinningRectangle2 animFrameLength1 = "▖▘▝▗"
	spinnerAnimShapes             animFrameLength1 = "┤┘┴└├┌┬┐"
	spinnerAnimMovingRectangle    animFrameLength1 = "▉▊▋▌▍▎▏▎▍▌▋▊▉"
	spinnerAnimPulsatingBar       animFrameLength1 = " ▁▂▃▄▅▆▇█▇▆▅▄▃▁"
	spinnerAnimWut                animFrameLength1 = "◡⊙◠⊙"
	spinnerAnimToggle             animFrameLength1 = "⊶⊷"
	spinnerAnimDots12             animFrameLength2 = "⢀⠀⡀⠀⠄⠀⢂⠀⡂⠀⠅⠀⢃⠀⡃⠀⠍⠀⢋⠀⡋⠀⠍⠁⢋⠁⡋⠁⠍⠉⠋⠉⠋⠉⠉⠙⠉⠙⠉⠩⠈⢙⠈⡙⢈⠩⡀⢙⠄⡙⢂⠩⡂⢘⠅⡘⢃⠨⡃⢐⠍⡐⢋⠠⡋⢀⠍⡁⢋⠁⡋⠁⠍⠉⠋⠉⠋⠉⠉⠙⠉⠙⠉⠩⠈⢙⠈⡙⠈⠩⠀⢙⠀⡙⠀⠩⠀⢘⠀⡘⠀⠨⠀⢐⠀⡐⠀⠠⠀⢀⠀⡀"
)

type Spinner struct {
	frames      []string
	started     time.Time
	speed       time.Duration
	needsUpdate *time.Ticker
}

func newSpinner(anim AnimFrameGetter) *Spinner {
	speed := 250 * time.Millisecond

	return &Spinner{
		frames:      anim.GetFrames(),
		started:     time.Now(),
		speed:       speed,
		needsUpdate: time.NewTicker(speed),
	}
}

func (s *Spinner) Get() string {
	animFrameIdx := int(time.Since(s.started)/s.speed) % len(s.frames)

	return s.frames[animFrameIdx]
}

func ProgressBar(pct int, barLength int, theme ProgressBarTheme) string {
	r := make([]rune, barLength)

	ratio := float64(barLength) * float64(pct) / 100.0

	for i := 0; i < barLength; i++ {
		ch := theme.Vacant
		if float64(i+1) <= ratio {
			ch = theme.Filled
		}

		r[i] = ch
	}

	return string(r)
}

type ProgressBarTheme struct {
	Filled rune
	Vacant rune
}

func ProgressBarDefaultTheme() ProgressBarTheme {
	return ProgressBarTheme{'█', '░'}
}

func ProgressBarCirclesTheme() ProgressBarTheme {
	return ProgressBarTheme{'⬤', '○'}
}
