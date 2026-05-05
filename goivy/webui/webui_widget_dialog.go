// Mechanical port of widget_dialog.py.
//
// Python source: ~/ivy/pyivy/ivy/ivy/widget_dialog.py (23 lines)
//
//   class DialogWidget(widgets.FlexBox):
//       _view_module = Unicode('nbextensions/ivy/js/widget_dialog', sync=True)
//       _view_name = Unicode('DialogView', sync=True)
//       options = Any(sync=True, doc="...")
//       title = Unicode(sync=True, doc="...")
//
//       def __init__(self, title, options, **kwargs):
//           kwargs['title'] = title
//           kwargs['options'] = options
//           super(DialogWidget, self).__init__(**kwargs)
//
// In Python this is an IPython FlexBox subclass that wraps a jQuery UI
// dialog. Go has no IPython, so the port is a data-carrier struct that the
// webui front end serializes to its JSON protocol.

package webui

// DialogWidget mirrors widget_dialog.py:DialogWidget. Its Title overrides
// any title set inside Options, and Options carries the kwargs that are
// passed to the front-end jQuery UI dialog/dialogExtend (notably 'height'
// and 'width', both of which support the 'max' value).
type DialogWidget struct {
	// Trait fields (Python class attributes)
	ViewModule string
	ViewName   string

	// Constructor fields
	Title    string
	Options  map[string]any
	Children []Widget

	// FlexBox kwargs the Python __init__ forwards via **kwargs.
	// In Python these come from widgets.FlexBox; in Go we expose only
	// the few that the widget_analysis_session port actually sets.
	Orientation string
	OverflowX   string
	OverflowY   string
}

// NewDialogWidget mirrors DialogWidget.__init__(self, title, options, **kwargs).
func NewDialogWidget(title string, options map[string]any) *DialogWidget {
	return &DialogWidget{
		ViewModule: "nbextensions/ivy/js/widget_dialog",
		ViewName:   "DialogView",
		Title:      title,
		Options:    options,
	}
}

func (*DialogWidget) widget() {}
