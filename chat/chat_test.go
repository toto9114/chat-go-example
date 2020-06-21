package chat

import (
	"fmt"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGinkgo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test functions Suite")
}

var _ = Describe("Chatting Test", func() {
	var (
		err error
		// response *http.Response
		ws *websocket.Conn
	)

	JustAfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			fmt.Printf("Collecting diags just after failed test in %s\n", CurrentGinkgoTestDescription().TestText)
		}
		ws.Close()
	})

	Describe("Test Websocket", func() {

		Context("Success Connect Websocket", func() {
			BeforeEach(func() {
				u := "ws://0.0.0.0:1213/chat/1/"
				ws, _, err = websocket.DefaultDialer.Dial(u, nil)
			})
			It("success connect", func() {
				By("start")
				Expect(err).To(BeNil())
				By("end")
			})
			It("should not error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("Success Send Message", func() {
			var (
				message []byte
				testMsg string
			)

			testMsg = "테스트입니다."
			BeforeEach(func() {

				interrupt := make(chan os.Signal, 1)
				signal.Notify(interrupt, os.Interrupt)

				u := "ws://0.0.0.0:1213/chat/1/"
				ws, _, err = websocket.DefaultDialer.Dial(u, nil)
				done := make(chan struct{})

				go func() {
					defer close(done)
					for {
						_, message, err = ws.ReadMessage()
						if err != nil {
							return
						}
						return
					}
				}()

				go func() {
					time.Sleep(2 * time.Second)
					ws.WriteMessage(websocket.TextMessage, []byte(`{"access_token": "Bearer UF3n6ayQCF34Q61VjE7L8OeRuHeugE", "message": "`+testMsg+`"}`))
				}()

				for {
					select {
					case <-done:
						return
					case <-interrupt:
						// Cleanly close the connection by sending a close message and then
						// waiting (with timeout) for the server to close the connection.
						err := ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
						if err != nil {
							return
						}
						select {
						case <-done:
						case <-time.After(time.Second):
						}
						return
					}
				}

			})

			It("success send message", func() {
				By("Start")
				Expect(err).To(BeNil())
				Expect(string(message)).To(Equal(testMsg))
				By("End")
			})
		})
	})
})
