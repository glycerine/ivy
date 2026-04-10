// Mechanical port of widget_modal_messages.py.
//
// Python source: ~/ivy/pyivy/ivy/ivy/widget_modal_messages.py (25 lines)
//
//   class ModalMessagesWidget(widgets.Widget):
//       _view_module = Unicode('nbextensions/ivy/js/widget_modal_messages', sync=True)
//       _view_name = Unicode('ModalMessagesView', sync=True)
//
//       def new_message(self, title, body):
//           """
//           Send a new message, title and body are strings
//           """
//           self.send({
//               "method": "new_message",
//               "title": title,
//               "body": body,
//           })
//
// In Python this widget exposes a single new_message() method that pushes
// a (title, body) pair to the front end via IPython's send() RPC. The Go
// port stores the messages and exposes them for the webui server to flush.

package webui

// ModalMessagesWidget mirrors widget_modal_messages.py:ModalMessagesWidget.
// It accumulates (title, body) message pairs that the front end displays
// as modal dialogs.
type ModalMessagesWidget struct {
	// Trait fields (Python class attributes)
	ViewModule string
	ViewName   string

	// Pending messages, in insertion order. Each entry corresponds to one
	// Python self.send({"method":"new_message", ...}) call.
	Messages []ModalMessage
}

// ModalMessage is one (title, body) pair queued for display.
type ModalMessage struct {
	Title string
	Body  string
}

// NewModalMessagesWidget constructs a fresh widget with no pending messages.
func NewModalMessagesWidget() *ModalMessagesWidget {
	return &ModalMessagesWidget{
		ViewModule: "nbextensions/ivy/js/widget_modal_messages",
		ViewName:   "ModalMessagesView",
	}
}

// NewMessage mirrors widget_modal_messages.py:new_message(self, title, body).
// It enqueues a new message; the webui server is responsible for flushing
// these to the front end via the JSON protocol.
func (m *ModalMessagesWidget) NewMessage(title, body string) {
	m.Messages = append(m.Messages, ModalMessage{Title: title, Body: body})
}

func (*ModalMessagesWidget) widget() {}
