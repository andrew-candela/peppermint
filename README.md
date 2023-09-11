
# UDPM - peer to peer messaging

This app provides a mechanism to communicate with other folks on the internet.
Messages are encrypted using a hybrid method - See the Security section.

## Security

Each message's content is signed with the user's private RSA key.
Before sending the message to each recipient, a new, random AES key
is used to encrypt the message content.
Then the AES key used to encrypt the message is itself encrypted
with the recipient's public RSA key.
The encrypted message is sent to the recipient along with the
signiature and the encrypted AES key.

For those of you serious about anonymity, know that this system only
obscures the content of your messages, not the members of the conversation.
This is way better than using Facebook or Google to message people,
but your ISP will probably have logs of who you've been sending packets to.
If you live in a country that likes to keep track of folks,
you might want to try something like TOR or a VPN (these are not perfect either).

## UDP mode

For a truly peer-to-peer experience, you can use UDP mode which requires you
to set up port forwarding on your local network.

## Web mode

Setting up port forwarding will not be feasible, or even possible in many cases.
If you cannot set up port forwarding, then you can use the WEB mode to communicate.
WEB mode requires one person in your group to host a UDPM server.
Messages are encrypted on each user's machine then sent to the server where
they are sent to each member of the group via websockets.

## User Interface

This is just here to remind myself of the intended architecture.
If you're reading this, you should probably think about spending your
time better.

```text
  user types a message
    user knows all recipients of the message
    message is signed
    for each recip:
      message is encrypted
      message is sent:
        message is serialized (message -> pbmessage)
        serialized message (pbmessage) is split into packets
        packets are sent to recipient
          WEB - publish endpoint with public key as user identifier
          UDP - udp connection via host:port
    user is notified of successful sends

  User listens to messages
    set up a goroutine for each member of the group
      listens to a channel associated with the group member
      knows the metadata of the user associated to the routine
      has a mutex for printing to stdOut so that we don't mix messages
      
      deserializes each packet
      combines all packets if there are many
      decrypts the AES key
      decrypts the message
      checks the signature
      prints the message to stdout
    Set up a multiplexer goroutine to assign messages to the listeners

    Listen loop:
      WEB
        register with the server as a consumer and establish websocket
        messages from websocket are sent to multiplexer
      UDP
        Listen for UDP messages on a given port
        messages from connections are sent to multiplexer
    multiplexer
      looks at origin (public key) and assigns to the right listener
      no message validation here

  Transports
    WEB

    UDP
```