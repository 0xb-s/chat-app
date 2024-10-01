package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	SERVER_HOST = "localhost"
	SERVER_PORT = "6000"
	SERVER_TYPE = "tcp"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading username:", err)
		return
	}
	username = strings.TrimSpace(username)
	if username == "" {
		fmt.Println("Username cannot be empty.")
		return
	}

	conn, err := net.Dial(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Send username to server
	_, err = fmt.Fprintln(conn, username)
	if err != nil {
		fmt.Println("Error sending username:", err)
		return
	}

	app := tview.NewApplication()

	chatView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWrap(true)
	chatView.SetBorder(true).SetTitle(" Chat ")

	chatView.SetTextColor(tcell.ColorWhite)
	chatView.SetBackgroundColor(tcell.ColorBlack)

	inputField := tview.NewInputField().
		SetLabel("Message: ")
	inputField.SetFieldWidth(0)

	inputField.SetFieldTextColor(tcell.ColorWhite)
	inputField.SetLabelColor(tcell.ColorGreen)
	inputField.SetBackgroundColor(tcell.ColorBlack)
	inputField.SetFieldBackgroundColor(tcell.ColorBlue)

	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			message := strings.TrimSpace(inputField.GetText())
			if message != "" {
				// Send message to server
				_, err := fmt.Fprintln(conn, message)
				if err != nil {
					// Display error in chat view
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(chatView, "[red]Error sending message: %v\n", err)
					})
				}

				inputField.SetText("")
			}
		}
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(chatView, 0, 1, false).
		AddItem(inputField, 3, 1, true)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			text := scanner.Text()
			app.QueueUpdateDraw(func() {
				fmt.Fprintln(chatView, text)
			})
		}
		if err := scanner.Err(); err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(chatView, "[red]Error reading from server: %v\n", err)
			})
		} else {
			app.QueueUpdateDraw(func() {
				fmt.Fprintln(chatView, "[red]Disconnected from server.")
			})
		}
	}()

	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		fmt.Println("Error running application:", err)
	}

	wg.Wait()
}
