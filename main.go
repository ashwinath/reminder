package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ashwinath/reminder/db"
)

func printTableHeader() {
	fmt.Printf("%-5s | %-20s | %-30s | %s\n", "ID", "Description", "URL", "Status")
}

func printReminder(r interface {
	GetID() int64
	GetDescription() string
	GetURL() string
	GetStatus() string
}) {
	fmt.Printf("%-5d | %-20s | %-30s | %s\n", r.GetID(), r.GetDescription(), r.GetURL(), r.GetStatus())
}

type ReminderRow struct {
	ID          int64
	Description string
	URL         string
	Status      string
}

func (r ReminderRow) GetID() int64          { return r.ID }
func (r ReminderRow) GetDescription() string { return r.Description }
func (r ReminderRow) GetURL() string         { return r.URL }
func (r ReminderRow) GetStatus() string      { return r.Status }

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	database, err := db.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	command := os.Args[1]

	switch command {
	case "add":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Usage: reminder add '<description>' '<url>'\n")
			os.Exit(1)
		}
		cmdAdd(database, os.Args[2], os.Args[3])
	case "complete":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: reminder complete <id1> [id2] ...\n")
			os.Exit(1)
		}
		cmdComplete(database, os.Args[2:])
	case "active":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: reminder active <id1> [id2] ...\n")
			os.Exit(1)
		}
		cmdActive(database, os.Args[2:])
	case "delete":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: reminder delete <id1> [id2] ...\n")
			os.Exit(1)
		}
		cmdDelete(database, os.Args[2:])
	case "get":
		status := "active"
		format := "table"
		topMessage := ""
		for _, arg := range os.Args[2:] {
			if arg == "all" {
				status = "all"
			} else if v, ok := strings.CutPrefix(arg, "--format="); ok {
				format = v
			} else if v, ok := strings.CutPrefix(arg, "--top-message="); ok {
				topMessage = v
			}
		}
		cmdGet(database, status, format, topMessage)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: reminder <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add <description> <url>    Add a new reminder")
	fmt.Println("  complete <id1> [id2] ...   Mark reminder(s) as completed")
	fmt.Println("  active <id1> [id2] ...   Mark reminder(s) as active")
	fmt.Println("  delete <id1> [id2] ...     Delete reminder(s)")
	fmt.Println("  get [all]                  Get reminders (default: active only)")
}

func cmdAdd(database *db.Database, description, url string) {
	reminder, err := database.Add(description, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Added reminder:")
	printTableHeader()
	printReminder(ReminderRow{
		ID:          reminder.ID,
		Description: reminder.Description,
		URL:         reminder.URL,
		Status:      reminder.Status,
	})
}

func cmdComplete(database *db.Database, args []string) {
	ids, err := parseIDs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	reminders, err := database.Complete(ids)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Completed reminder:")
	printTableHeader()
	for _, r := range reminders {
		printReminder(ReminderRow{
			ID:          r.ID,
			Description: r.Description,
			URL:         r.URL,
			Status:      r.Status,
		})
	}
}

func cmdActive(database *db.Database, args []string) {
	ids, err := parseIDs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	reminders, err := database.Activate(ids)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Activated reminder:")
	printTableHeader()
	for _, r := range reminders {
		printReminder(ReminderRow{
			ID:          r.ID,
			Description: r.Description,
			URL:         r.URL,
			Status:      r.Status,
		})
	}
}

func cmdDelete(database *db.Database, args []string) {
	ids, err := parseIDs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	reminders, err := database.Delete(ids)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Deleted reminder:")
	printTableHeader()
	for _, r := range reminders {
		printReminder(ReminderRow{
			ID:          r.ID,
			Description: r.Description,
			URL:         r.URL,
			Status:      "deleted",
		})
	}
}

func cmdGet(database *db.Database, status, format, topMessage string) {
	reminders, err := database.GetAll(status)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(reminders) == 0 {
		fmt.Println("No reminders found.")
		return
	}

	if format == "slack" {
		if topMessage != "" {
			fmt.Println(topMessage)
		}
		for _, r := range reminders {
			fmt.Printf("%s - %s\n", r.Description, r.URL)
		}
		return
	}

	printTableHeader()
	for _, r := range reminders {
		printReminder(ReminderRow{
			ID:          r.ID,
			Description: r.Description,
			URL:         r.URL,
			Status:      r.Status,
		})
	}
}

func parseIDs(args []string) ([]int64, error) {
	var ids []int64
	for _, arg := range args {
		id, err := strconv.ParseInt(arg, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ID: %s", arg)
		}
		ids = append(ids, id)
	}
	return ids, nil
}
