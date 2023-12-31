
# Peppermint - Peer to Peer Messaging in a Terminal

![Example](example.png)

This CLI tool provides a mechanism to communicate with other folks over the internet.

PPMT users can organize themselves and their contacts into Groups.
Groups are like channels in slack.
A message sent to the group is seen by all group members.

In order to facilitate messaging, someone must make a PPMT server available.
Clients subscribe to a group by initiating a websocket connection with the server.
Anyone on a network capable of accepting inbound connections can host a PPMT server.

## Install

Peppermint is an executable file with no dependencies.
Download the binary for your system from the
[releases](https://github.com/andrew-candela/peppermint/releases) page.

Then, unpack the tarbal (`tar -xf peppermint_release.tar.gz` on unix-like systems)
and move the binary to somewhere in your system's PATH.

## Usage

```bash
# Initialize your peppermint config
peppermint init

# After editing ~/.peppermint/config
# Run a peppermint server
peppermint host

# Listen for messages in a group
peppermint read -g your_group_name

# Write messages to a group
peppermint write -g your_group_name
```

## Encryption

Messages sent with PPMT are encrypted and decrypted on the clients.
The PPMT server has no way to decrypt the bytes that it forwards to subscribers.

PPMT uses hybrid encryption.
A new, random AES key is generated for each message/recipient pair.
The message content is encrypted with the AES key, and then the AES key
is encrypted with the public key of the message recipient.
