# CefEnum

## About
CefEnum is a tool designed to detect and enumerate CEFSharp based thick-clients, identifying .NET objects exposed to JavaScript for potential security research and red teaming. It helps security researchers and penetration testers quickly fingerprint CEFSharp instances and discover exploitable object bindings in .NET-based applications.
For a comprehensive breakdown of CefEnum's capabilities, usage scenarios, and attack vectors, please visit our detailed blog post: [CEFSharp Enumeration Methods With CefEnum.](https://blog.darkforge.io/cef/cefsharp/cefenum/thick-client/.net/2025/05/21/CefSharp-Enumeration-With-CefEnum.html)

## Installation
To install CefEnum, use the following command:
```
go install github.com/darkforge-labs/cefenum@latest
```

Ensure you have Go installed on your system. CefEnum will be installed to your $GOPATH/bin directory.

## Usage
Once installed, run CefEnum with the following command:
```
cefenum [flags]
```

Available Flags
```
-port string: Port to run the server on (default: "9090")
-wordlist string: Path to a custom wordlist file for fuzzing object names (default: built-in wordlist)
```

By default, CefEnum starts an HTTP listener on port 9090 and uses a built-in wordlist based on PortSwigger's param-miner.

## Example
To run CefEnum with default settings:
`cefenum`

To specify a custom port and wordlist:
`cefenum -port 8080 -wordlist ./custom_wordlist.txt`

## Interacting with the CEFSharp Client
Once CefEnum is running, it opens an interactive shell where you can send commands to a connected CEFSharp client. The following commands are available:

```
exit/quit: Exit the program
clear: Clear the console
fuzz: Bind common objects using the wordlist
detect: Detect if CefSharp is present in the client
brute: Brute-force object discovery [a-zA-Z]
bind <objectName>: Bind and Inspect a specific object's methods
help: Show the help message
```

## Notes
CefEnum delivers a wordlist to the client upon connection, which is used to fuzz for exposed .NET object names.
The tool communicates with the client over a WebSocket, allowing real-time interaction via the command-line interface.
Brute-forcing with the brute command is not recommended for object names longer than five characters due to performance limitations.

## Contributing
Contributions are welcome! If you have ideas for improving CefEnum, such as faster enumeration techniques or additional features, please submit a pull request or open an issue on this repository.

## License
This project is licensed under the MIT License. See the LICENSE file for details.

## Contact
For questions, feedback, or professional inquiries, contact us at DarkForge Labs via our website or submit an issue on this repository.
