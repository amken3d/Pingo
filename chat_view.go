// ┌──────────────────────────────────────────────────────────────────────┐
// │  chat_view.go — Custom Chat View with Selectable Text              │
// │                                                                    │
// │  Replaces the default ChatPanel.Layout() with a version that uses  │
// │  read-only giowidget.Editor for each message, enabling text        │
// │  selection and copy-to-clipboard.                                  │
// └──────────────────────────────────────────────────────────────────────┘
package main

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	giowidget "gioui.org/widget"

	"github.com/amken3d/immygo/ai"
	"github.com/amken3d/immygo/theme"
	"github.com/amken3d/immygo/ui"
)

// chatMsgEditors holds persistent read-only editors for chat messages.
// Grows dynamically as messages arrive.
var chatMsgEditors []*giowidget.Editor

// chatScrollList is a persistent scroll list for the chat messages area.
var chatScrollList = func() *giowidget.List {
	l := &giowidget.List{}
	l.Axis = layout.Vertical
	return l
}()

// ensureChatEditors grows the editor slice to match the message count.
func ensureChatEditors(n int) {
	for len(chatMsgEditors) < n {
		ed := &giowidget.Editor{}
		ed.ReadOnly = true
		ed.SingleLine = false
		chatMsgEditors = append(chatMsgEditors, ed)
	}
}

// selectableChatView returns a ui.View that renders the chat panel
// with selectable text in message bubbles.
func selectableChatView() ui.View {
	return ui.ViewFunc(func(gtx layout.Context, th *theme.Theme) layout.Dimensions {
		// Process input events (submit, send button click).
		for {
			ev, ok := chatPanel.Input.Update(gtx)
			if !ok {
				break
			}
			if _, ok := ev.(giowidget.SubmitEvent); ok {
				text := chatPanel.Input.Text()
				if text != "" {
					chatPanel.Input.SetText("")
					chatPanel.SendMessage(text)
				}
			}
		}
		if chatPanel.SendBtn.Clicked(gtx) {
			text := chatPanel.Input.Text()
			if text != "" {
				chatPanel.Input.SetText("")
				chatPanel.SendMessage(text)
			}
		}

		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Input bar at top.
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layoutChatInput(gtx, th)
			}),
			// Scrollable message area fills remaining space.
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				// Messages is public; reads from the UI goroutine are safe
				// as the original Layout() also reads directly in the same context.
				msgs := make([]ai.Message, len(chatPanel.Messages))
				copy(msgs, chatPanel.Messages)

				n := len(msgs)
				ensureChatEditors(n)

				return chatScrollList.Layout(gtx, n, func(gtx layout.Context, index int) layout.Dimensions {
					// Newest messages at top.
					msg := msgs[n-1-index]
					editorIdx := n - 1 - index
					ed := chatMsgEditors[editorIdx]

					// Sync editor text if message content changed.
					if ed.Text() != msg.Content {
						ed.SetText(msg.Content)
					}

					inset := layout.Inset{
						Top:    unit.Dp(4),
						Bottom: unit.Dp(4),
						Left:   unit.Dp(8),
						Right:  unit.Dp(8),
					}
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						return layoutSelectableBubble(gtx, th, ed, msg.Role == ai.RoleUser)
					})
				})
			}),
		)
	})
}

// layoutSelectableBubble renders a chat bubble with a read-only selectable editor.
func layoutSelectableBubble(gtx layout.Context, th *theme.Theme, ed *giowidget.Editor, isUser bool) layout.Dimensions {
	maxW := gtx.Constraints.Max.X * 3 / 4

	var bgColor, fgColor color.NRGBA
	radius := 12

	if isUser {
		bgColor = th.Palette.Primary
		fgColor = th.Palette.OnPrimary
	} else {
		bgColor = th.Palette.SurfaceVariant
		fgColor = th.Palette.OnSurface
	}

	return layout.Flex{
		Axis:    layout.Horizontal,
		Spacing: layout.SpaceStart,
	}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = maxW

			return layout.Stack{}.Layout(gtx,
				layout.Expanded(func(gtx layout.Context) layout.Dimensions {
					size := gtx.Constraints.Min
					rr := clip.UniformRRect(image.Rectangle{Max: size}, radius)
					defer rr.Push(gtx.Ops).Pop()
					paint.ColorOp{Color: bgColor}.Add(gtx.Ops)
					paint.PaintOp{}.Add(gtx.Ops)
					return layout.Dimensions{Size: size}
				}),
				layout.Stacked(func(gtx layout.Context) layout.Dimensions {
					inset := layout.UniformInset(unit.Dp(12))
					return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						selectColor := theme.WithAlpha(th.Palette.Primary, 60)
						return ed.Layout(gtx, th.Shaper, font.Font{}, th.Typo.BodyMedium.Size, colorMaterial(gtx.Ops, fgColor), colorMaterial(gtx.Ops, selectColor))
					})
				}),
			)
		}),
	)
}

// colorMaterial records a paint color on the given ops and returns it as a CallOp.
func colorMaterial(ops *op.Ops, c color.NRGBA) op.CallOp {
	m := op.Record(ops)
	paint.ColorOp{Color: c}.Add(ops)
	return m.Stop()
}

// layoutChatInput renders the chat input bar (editor + send button).
func layoutChatInput(gtx layout.Context, th *theme.Theme) layout.Dimensions {
	return layout.Stack{}.Layout(gtx,
		layout.Expanded(func(gtx layout.Context) layout.Dimensions {
			size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Min.Y}
			rr := clip.UniformRRect(image.Rectangle{Max: size}, 0)
			defer rr.Push(gtx.Ops).Pop()
			paint.ColorOp{Color: th.Palette.Surface}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			// Bottom border.
			borderSize := image.Point{X: size.X, Y: 1}
			rr2 := clip.Rect(image.Rectangle{Max: borderSize})
			defer rr2.Push(gtx.Ops).Pop()
			paint.ColorOp{Color: th.Palette.OutlineVariant}.Add(gtx.Ops)
			paint.PaintOp{}.Add(gtx.Ops)
			return layout.Dimensions{Size: size}
		}),
		layout.Stacked(func(gtx layout.Context) layout.Dimensions {
			inset := layout.Inset{
				Top:    unit.Dp(8),
				Bottom: unit.Dp(8),
				Left:   unit.Dp(12),
				Right:  unit.Dp(12),
			}
			return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				selectColor := theme.WithAlpha(th.Palette.Primary, 60)
				return layout.Flex{
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return chatPanel.Input.Layout(gtx, th.Shaper, font.Font{}, th.Typo.BodyMedium.Size, colorMaterial(gtx.Ops, th.Palette.OnSurface), colorMaterial(gtx.Ops, selectColor))
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Spacer{Width: unit.Dp(8)}.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						size := image.Point{X: gtx.Dp(unit.Dp(32)), Y: gtx.Dp(unit.Dp(32))}
						return chatPanel.SendBtn.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
							bgCol := th.Palette.Primary
							rr := clip.UniformRRect(image.Rectangle{Max: size}, size.X/2)
							defer rr.Push(gtx.Ops).Pop()
							paint.ColorOp{Color: bgCol}.Add(gtx.Ops)
							paint.PaintOp{}.Add(gtx.Ops)

							// Arrow icon.
							arrowOff := op.Offset(image.Pt(size.X/2-4, size.Y/2-5)).Push(gtx.Ops)
							var p clip.Path
							p.Begin(gtx.Ops)
							p.MoveTo(f32.Pt(0, 10))
							p.LineTo(f32.Pt(4, 0))
							p.LineTo(f32.Pt(8, 10))
							defer clip.Stroke{Path: p.End(), Width: 2}.Op().Push(gtx.Ops).Pop()
							paint.ColorOp{Color: th.Palette.OnPrimary}.Add(gtx.Ops)
							paint.PaintOp{}.Add(gtx.Ops)
							arrowOff.Pop()

							return layout.Dimensions{Size: size}
						})
					}),
				)
			})
		}),
	)
}
