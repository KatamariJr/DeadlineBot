# DeadlineBot: A Discord bot to remind you when things are due.

Make a file called `bot.json` in the same directory as where this program runs. The JSON document allows you to set values for the bot, using the following keys:
```
"discord-token" (string): The token for your discord bot that this program controls.
"channel-id" (string): The channel id where you want your bot to send the reminder message.
```

Run the program and give the first argument as a string representing the deadline in this format: `DD MON YY HH:MM:SS TZ`. For example, `29 Oct 20 18:00 EDT`.