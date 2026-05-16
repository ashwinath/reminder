# Reminder CLI

## Architecture
- golang
- sqlite

## Requirements
This app is a cli app that keeps a list of reminders. We can save the sqlite db in ${HOME}/.reminder/db.sqlite

A reminder has the following attributes:
- id: number
- status: active/completed
- description: just plain text
- URL: url relevant to the reminder
- created date
- updated date

### Adding Reminders

To add a reminder, the user just has to issue the following command.
`reminder add '$description' '$url'`

And it should return the following:
```
Added reminder:
ID   | Description    | URL    | status
<id> | <$description> | <$url> | active
```

### Marking reminder as completed

To mark a reminder as complete, the user just has to issue the following command. You can mark multiple IDs as complete at once
`reminder complete <$id1> <$id2>`

And it should return the following:

```
Completed reminder:
ID   | Description    | URL    | status
<$id1> | <$description> | <$url> | completed
<$id2> | <$description> | <$url> | completed
```

### Marking reminder as active

To mark a reminder as active, the user just has to issue the following command. You can mark multiple IDs as active at once
`reminder active <$id1> <$id2>`

And it should return the following:

```
Completed reminder:
ID   | Description    | URL    | status
<$id1> | <$description> | <$url> | active
<$id2> | <$description> | <$url> | active
```

### Deleting reminders

To delete a reminder, the user just has to issue the following command. You can delete multiple IDs at once.
`reminder delete <$id1> <$id2>`

And it should return the following but the data is actually deleted:

```
Deleted reminder:
ID   | Description    | URL    | status
<$id1> | <$description> | <$url> | deleted
<$id2> | <$description> | <$url> | deleted
```

### Getting reminders
By default we will only return active reminders:
`reminder get`

And it should return the following:
```
ID   | Description    | URL    | status
<$id1> | <$description> | <$url> | active
<$id2> | <$description> | <$url> | active
```

If we want to get all statuses, we can do
`reminder get all`
```
ID   | Description    | URL    | status
<$id1> | <$description> | <$url> | active
<$id2> | <$description> | <$url> | active
<$id3> | <$description> | <$url> | completed
```

If we want it in a slack friendly format, we can do
`reminder get --format=slack` --top-message=<$top_message>

```
${top_message}
<$description1> - $url1
<$description2> - $url2
<$description3> - $url3
```
