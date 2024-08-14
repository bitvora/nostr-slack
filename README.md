# nostr-slack

This is a Go application that listens to Nostr events from a set of authors and posts them to a Slack channel via a webhook. The authors to follow are configured in a `authors.json` file.

## Features

- Connects to a Nostr relay and subscribes to events from specified authors.
- Posts formatted messages to a Slack channel.
- Authors are configurable via a JSON file.

## Prerequisites

- Go 1.16+ installed.
- A Nostr relay to connect to.
- A Slack webhook URL.

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/bitvora/nostr-slack
   cd nostr-slack
   ```

2. Set your .env variables

   ```sh
   cp .env.example .env
   ```

## Configuration

1. Create a JSON file named `authors.json` in the root of the project. This file will contain the authors you want to follow. Here’s an example:

   ```json
   [
     {
       "npub": "npub10pensatlcfwktnvjjw2dtem38n6rvw8g6fv73h84cuacxn4c28eqyfn34f",
       "name": "OpenSats",
       "link": "https://njump.me/npub10pensatlcfwktnvjjw2dtem38n6rvw8g6fv73h84cuacxn4c28eqyfn34f"
     },
     {
       "npub": "npub1hxjnw53mhghumt590kgd3fmqme8jzwwflyxesmm50nnapmqdzu7swqagw3",
       "name": "No BS Bitcoin",
       "link": "https://njump.me/npub1hxjnw53mhghumt590kgd3fmqme8jzwwflyxesmm50nnapmqdzu7swqagw3"
     },
     {
       "npub": "npub1w4dsvkv5hq73p4wm6gadpcxs6fwshcys44f5tnnzze2g3hfs2p0qn23vhw",
       "name": "Canadian Bitcoiners Podcast",
       "link": "https://njump.me/npub1w4dsvkv5hq73p4wm6gadpcxs6fwshcys44f5tnnzze2g3hfs2p0qn23vhw"
     }
   ]
   ```

## Usage

1. Run the application:

   ```sh
   go run main.go
   ```

2. The application will connect to the Nostr relay, subscribe to the authors specified in `authors.json`, and start listening for new events.

3. When a new event is detected, it will be posted to your Slack channel using the specified format.

## Customization

- **Authors:** Modify the `authors.json` file to add or remove authors you want to follow.
- **Slack Message Format:** Adjust the formatting of the Slack message in the `slackMessage` string in `main.go` to suit your needs.

## Contributing

Feel free to fork this repository and submit pull requests. All contributions are welcome!

## License

This project is licensed under the MIT License
