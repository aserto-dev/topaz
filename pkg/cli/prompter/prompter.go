package prompter

import (
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pkg/errors"

	"github.com/rivo/tview"
	"github.com/samber/lo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type prompt struct {
	app  *tview.Application
	form *tview.Form
	msg  proto.Message
	rc   chan bool
}

func New(msg proto.Message) *prompt {
	return &prompt{
		app:  tview.NewApplication(),
		form: tview.NewForm(),
		msg:  msg,
		rc:   make(chan bool, 1),
	}
}

func (f *prompt) Req() proto.Message {
	return f.msg
}

func (f *prompt) Show() error {
	if err := f.init(); err != nil {
		return err
	}
	defer close(f.rc)

	if err := f.app.SetRoot(f.form, false).EnableMouse(true).Run(); err != nil {
		return err
	}

	rc := <-f.rc
	if !rc {
		return ErrCancelled
	}
	return nil
}

func (f *prompt) init() error {
	md := f.msg.ProtoReflect().Descriptor()
	name := md.FullName().Name()

	f.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC || event.Key() == tcell.KeyCtrlC {
			f.rc <- false
			f.app.Stop()
			return nil
		}
		return event
	})

	f.form.SetBorder(true)
	f.form.SetTitle(string(name)).SetTitleAlign(tview.AlignLeft)

	_ = f.addFields(f.msg, f.msg.ProtoReflect().Descriptor(), []string{})

	f.form.AddButton("Submit", func() {
		f.rc <- true
		f.app.Stop()
	})

	f.form.AddButton("Cancel", func() {
		f.rc <- false
		f.app.Stop()
	})

	c := f.form.GetFormItemCount()
	maxLabel := 0
	for i := 0; i < c; i++ {
		maxLabel = max(maxLabel, len(f.form.GetFormItem(i).GetLabel()))
	}
	width := maxLabel + 6 + 64
	f.form.SetRect(0, 0, width, ((c+3)*2)-1)

	return nil
}

var (
	ErrUnknownKind = status.Error(codes.Unknown, "unknown kind")
	ErrCancelled   = status.Error(codes.Aborted, "canceled")
)

func (f *prompt) addFields(msg proto.Message, md protoreflect.MessageDescriptor, parent []string) error {
	fields := md.Fields()

	for i := 0; i < fields.Len(); i++ {
		c := i
		path := append(parent, fields.Get(i).TextName()) //nolint: gocritic
		fieldName := strings.Join(path, ".")
		fd := fields.Get(i)

		switch fields.Get(i).Kind() { //nolint: exhaustive
		case protoreflect.StringKind:
			if fields.Get(i).IsList() {
				f.form.AddInputField(fieldName, "[ ]", 64, nil, func(s string) {
					_ = f.setProps(msg.ProtoReflect(), fields.Get(c), s)
				})
				continue
			}

			v := msg.ProtoReflect().Get(fd)
			f.form.AddInputField(fieldName, v.String(), 64, nil, func(s string) {
				_ = f.setField(msg.ProtoReflect(), fields.Get(c), s)
			})

		case protoreflect.Int32Kind:
			v := msg.ProtoReflect().Get(fd)
			f.form.AddInputField(fieldName, v.String(), 10, nil, func(s string) {
				_ = f.setField(msg.ProtoReflect(), fields.Get(c), s)
			})

		case protoreflect.Int64Kind:
			v := msg.ProtoReflect().Get(fd)
			f.form.AddInputField(fieldName, v.String(), 10, nil, func(s string) {
				_ = f.setField(msg.ProtoReflect(), fields.Get(c), s)
			})

		case protoreflect.BoolKind:
			v := msg.ProtoReflect().Get(fd)
			f.form.AddCheckbox(fieldName, v.Bool(), func(b bool) {
				_ = f.setField(msg.ProtoReflect(), fields.Get(c), lo.Ternary(b, "true", "false"))
			})

		case protoreflect.EnumKind:
			options := []string{}
			lookup := map[string]int32{}
			for v := 0; v < fields.Get(i).Enum().Values().Len(); v++ {
				name := string(fields.Get(i).Enum().Values().Get(v).Name())
				number := int32(fields.Get(i).Enum().Values().Get(v).Number())
				options = append(options, name)
				lookup[name] = number
			}

			f.form.AddDropDown(fieldName, options, 0, func(opt string, index int) {
				number, ok := lookup[opt]
				if ok {
					_ = f.setEnum(msg.ProtoReflect(), fields.Get(c), number)
				}
			})

		case protoreflect.MessageKind:
			if strings.HasSuffix(string(fd.FullName()), ".properties") {
				f.form.AddInputField(fieldName, "{ }", 64, nil, func(s string) {
					_ = f.setProps(msg.ProtoReflect(), fields.Get(c), s)
				})
				continue
			}

			if strings.HasSuffix(string(fd.FullName()), ".resource_context") {
				f.form.AddInputField(fieldName, "{ }", 64, nil, func(s string) {
					_ = f.setProps(msg.ProtoReflect(), fields.Get(c), s)
				})
				continue
			}

			if strings.HasSuffix(string(fd.FullName()), ".created_at") {
				f.form.AddInputField(fieldName, "", 64, nil, func(s string) {
					_ = f.setTimestamp(msg.ProtoReflect(), fields.Get(c), s)
				})
				continue
			}

			if strings.HasSuffix(string(fd.FullName()), ".updated_at") {
				f.form.AddInputField(fieldName, "", 64, nil, func(s string) {
					_ = f.setTimestamp(msg.ProtoReflect(), fields.Get(c), s)
				})
				continue
			}

			cm := msg.ProtoReflect().Get(fd).Message()
			_ = f.addFields(cm.Interface(), fields.Get(c).Message(), path)

		default:
			return errors.Wrapf(ErrUnknownKind, "%s", fields.Get(i).FullName())
		}
	}

	return nil
}

func (f *prompt) setField(msg protoreflect.Message, fd protoreflect.FieldDescriptor, value string) error {
	switch fd.Kind() { //nolint: exhaustive
	case protoreflect.BoolKind:
		b, err := strconv.ParseBool(lo.Ternary(value == "", "false", value))
		if err != nil {
			return err
		}

		msg.Set(fd, protoreflect.ValueOfBool(b))

	case protoreflect.Int32Kind:
		i, err := strconv.ParseInt(lo.Ternary(value == "", "false", value), 10, 32)
		if err != nil {
			return err
		}

		msg.Set(fd, protoreflect.ValueOfInt32(int32(i)))

	case protoreflect.Int64Kind:
		i, err := strconv.ParseInt(lo.Ternary(value == "", "false", value), 10, 64)
		if err != nil {
			return err
		}

		msg.Set(fd, protoreflect.ValueOfInt64(i))

	case protoreflect.StringKind:
		msg.Set(fd, protoreflect.ValueOfString(value))

	default:
		return errors.Wrapf(ErrUnknownKind, "%s", fd.Kind().String())
	}

	return nil
}

func (f *prompt) setProps(msg protoreflect.Message, fd protoreflect.FieldDescriptor, value string) error {
	pbs := &structpb.Struct{Fields: make(map[string]*structpb.Value)}
	if err := pbs.UnmarshalJSON([]byte(value)); err != nil {
		return err
	}
	msg.Set(fd, protoreflect.ValueOfMessage(pbs.ProtoReflect()))
	return nil
}

func (f *prompt) setTimestamp(msg protoreflect.Message, fd protoreflect.FieldDescriptor, value string) error {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return err
	}
	ts := timestamppb.New(t)
	msg.Set(fd, protoreflect.ValueOfMessage(ts.ProtoReflect()))
	return nil
}

func (f *prompt) setEnum(msg protoreflect.Message, fd protoreflect.FieldDescriptor, index int32) error {
	msg.Set(fd, protoreflect.ValueOfEnum(protoreflect.EnumNumber(index)))
	return nil
}
