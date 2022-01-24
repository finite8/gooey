package core

import (
	"context"
	"html/template"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/ntaylor-barnett/gooey/register"
	"github.com/sirupsen/logrus"
)

var textStreamTemplate = template.Must(template.New("stream").Parse(`<pre id="output"></pre>
<script>
   
    var output = document.getElementById("output");
    var socket = new WebSocket("ws://{{.StreamURL}}");

    socket.onopen = function () {
        output.innerHTML += "Status: Connected\n";
    };

    socket.onmessage = function (e) {
        output.innerHTML += "Server: " + e.data + "\n";
    };

    function send() {
        socket.send(input.value);
        input.value = "";
    }
</script>`))

// TextStreamComponent is designed to allow an implementation to keep sending updates
type TextStreamComponent struct {
	ComponentBase
	streamWorker func(context.Context, chan<- string) error
	streamPage   register.Page
}

func NewStreamComponent(f func(context.Context, chan<- string) error) *TextStreamComponent {
	tsr := &TextStreamComponent{
		streamWorker: f,
	}
	return tsr

}

func (tc *TextStreamComponent) WriteContent(ctx register.PageContext, w PageWriter) {
	strmUrl := ctx.GetPageUrl(tc.streamPage)
	var tmplLoad struct{ StreamURL template.URL }
	tmplLoad.StreamURL = template.URL(strmUrl.Host + strmUrl.Path)
	textStreamTemplate.Execute(w, tmplLoad)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (tc *TextStreamComponent) OnRegister(ctx register.Registerer) {
	strmPage := register.NewAPIPage("stream", func(pctx register.PageContext, rw http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(rw, r, nil) // error ignored for sake of simplicity
		var wg sync.WaitGroup
		ctx, canceller := context.WithCancel(context.Background())
		origCancel := canceller
		canceller = func() {
			origCancel()
		}
		txtChan := make(chan string, 50)

		wg.Add(1)
		go func() {
			defer canceller()
			defer close(txtChan)

			// worker to handle the execution of the function given to us
			err := tc.streamWorker(ctx, txtChan)
			if err != nil {
				logrus.Error(err)
			}
		}()
		wg.Add(1)
		go func() {
			defer canceller()
			// worker to handle responding to client side events
			for {
				msgType, data, err := conn.ReadMessage()
				if err != nil {
					return
				}
				if msgType == websocket.CloseMessage {
					return
				}
				_ = data
			}
		}()

		wg.Add(1)
		go func() {
			// worker thread to handle pushing buffered messages to the browser
			defer canceller()
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-txtChan:
					if !ok {
						// channel is closed
						conn.Close()
						return
					}
					err := conn.WriteMessage(websocket.TextMessage, []byte(msg))

					if err != nil {
						return
					}
				}
			}
		}()
		wg.Wait()
	})
	ctx.RegisterPrivateSubPage("stream", strmPage)
	tc.streamPage = strmPage
}
