// example.go - Complete example program
package ui

import (
	"fmt"
	"log"
	"strings"

	"github.com/robbiew/retrograde/internal/jam"
)

func main() {
	// Example 1: Creating a new message base and posting a local message
	fmt.Println("=== Example 1: Creating and posting to a local message base ===")

	base, err := jam.Open("./msgbase/local")
	if err != nil {
		log.Fatal(err)
	}
	defer base.Close()

	// Post a local message
	msg := jam.NewMessage().
		From("SysOp").
		To("All").
		Subject("Welcome to the BBS!").
		Text("This is a test message in the local message base.\n\nEnjoy your stay!").
		Local().
		Build()

	msgNum, err := base.WriteMessage(msg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Posted message #%d\n\n", msgNum)

	// Example 2: Creating a netmail message
	fmt.Println("=== Example 2: Creating a netmail message ===")

	netmailBase, err := jam.Open("./msgbase/netmail")
	if err != nil {
		log.Fatal(err)
	}
	defer netmailBase.Close()

	netmail := jam.NewNetmail().
		From("John Doe", "1:123/456").
		To("Jane Smith", "1:123/789").
		Subject("Test Netmail").
		Text("This is a test netmail message.").
		Private().
		Build()

	msgNum, err = netmailBase.WriteMessage(netmail)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Posted netmail #%d\n\n", msgNum)

	// Example 3: Creating an echomail message
	fmt.Println("=== Example 3: Creating an echomail message ===")

	echoBase, err := jam.Open("./msgbase/echo_general")
	if err != nil {
		log.Fatal(err)
	}
	defer echoBase.Close()

	echo := jam.NewEchomail("FIDOTEST").
		From("SysOp", "1:123/456").
		To("All").
		Subject("Test Echo Message").
		Text("This is a test echomail message.").
		AddSeenBy("1:123/456").
		AddPath("1:123/456").
		AddArea().
		Build()

	msgNum, err = echoBase.WriteMessage(echo)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Posted echomail #%d\n\n", msgNum)

	// Example 4: Reading and displaying messages
	fmt.Println("=== Example 4: Reading messages ===")

	count, _ := base.GetMessageCount()
	for i := 1; i <= count; i++ {
		msg, err := base.ReadMessage(i)
		if err != nil {
			continue
		}

		fmt.Printf("Message #%d\n", i)
		fmt.Printf("From: %s\n", msg.From)
		fmt.Printf("To: %s\n", msg.To)
		fmt.Printf("Subject: %s\n", msg.Subject)
		fmt.Printf("Date: %s\n", msg.DateTime.Format("2006-01-02 15:04:05"))
		fmt.Printf("Text:\n%s\n", msg.Text)
		fmt.Println(strings.Repeat("-", 70))
	}

	// Example 5: Scanning for specific messages
	fmt.Println("\n=== Example 5: Scanning for messages ===")

	scanner := jam.NewScanner(base)
	messages, err := scanner.ScanByFrom("SysOp")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d messages from SysOp\n\n", len(messages))

	// Example 6: Message base maintenance
	fmt.Println("=== Example 6: Message base statistics ===")

	maint := jam.NewMaintenance(base)
	stats, err := maint.GetStatistics()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(stats)
}
