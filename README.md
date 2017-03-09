# GoMinewrap [![MIT license](https://img.shields.io/badge/license-MIT-blue.svg)]() [![Chat on discord](https://img.shields.io/badge/chat%20on-discord-yellow.svg)](https://discord.gg/tae9mst) [![PayPal me!](https://img.shields.io/badge/PayPal-me-lightgrey.svg)](https://www.paypal.me/SuperPykkon)
GoMinewrap is a wrapper for your Minecraft server console. It comes with many new features that could possibly make server management easier. Released under MIT license.  
It allows you to control multiple server, switch between them, and even back them up. It includes a web interface with a modern console log with highlighting, the web interface supports multiple users too.  

## Prerequisites
* [Go](https://golang.org) 1.8

## Cloning and Building
1. `go get github.com/SuperPykkon/GoMinewrap`  
2. `cd $GOPATH/src/github.com/SuperPykkon/GoMinewrap`  
3. `go run *.go`  
Pre-compiled executables are available on the releases tab.

<<<<<<< HEAD
## Running
You will need to configure the configuration under `GoMinewrap/config.yml` to set the servers for the wrapper. You can manually set the config path using a flag. By default, the web server runs on port 25465. For help on wrapper commands, type !help on the command line.
=======
# How to use it?  
Firsly, you have the option to use the .go file or the executables. Currently, I've made executables for Windows and Linux which you can download here:  
  - Windows: https://github.com/SuperPykkon/GoMinewrap/files/830959/gominewrap-0.3-windows.zip  
  - Linux: https://github.com/SuperPykkon/GoMinewrap/files/830960/gominewrap-0.3-linux.zip  
  - Mac: https://github.com/SuperPykkon/GoMinewrap/files/830958/gominewrap-0.3-mac.zip  
  
  **NOTE**
  Please note that the executables for Linux and Mac is **not tested**. I don't have any of them and I can't get a VM either.  
  So, it is very much possible that it won't work. I recommend you to run mcs.go instead or just use Wine on the Windows executable.  
  A proper working version of these will be released soon!  
  
# How to use it? -- Webcon
  
When you download the zip, it comes with the executable and static folder. This is where all the html files for webcon are.  
You can also place the files else where, but you have to make sure to provide the new path on the config: *webcon -> dir*.  
  
To use webcon, you will first have to enable it on the config: *webcon -> enabled*. Don't forget to add a username and password for the login on the config: *webcon -> users*, you can add as many users you want.  
  
By default, webcon runs on localhost on the port 80, so http://127.0.0.1 or http://localhost should work. Or you can specify your own host and port on the config: *webcon -> host* and *webcon -> port*.  
  
# How to use it? -- Minecraft server
  
By default, the zip comes with a folder called "server", this is where your server files will go. You can place the executable anywhere, just make sure to set the path/to/serverFiles/ on the config: *server -> dir*.  
  
If you're placing the execuable in the same directory as spigot.jar, use *"."* on *server -> dir*  
By default, the startup script is **java -Xmx1G -jar spigot.jar**, you can change it on the config: *server -> startup*.

# All the flags
    1. --config [name]
       Name of config file (without extension)
       default: config
    
    2. --configDir [path/to/config/]
       Path to look for the config file in
       default: GoMinewrap/
 
  It's best to leave it as it is :P
  
# The config.yml file

    #
    #    GoMinewrap config file.
    #

    # Do not change.
    version: "0.3"
    server:
        # What's the name of your server?
        name: "Alephnull"
        # Enter the path to the folder with all your servers
        base: "servers/"
        # Enter the primary/main server here. CaSe SeNsItIvE
        primary: "hub"
        # Add all your servers here.
        servers:
            # The server's name.
            hub:
                # The server's name.
                # Enable or disable automatic server startup when the program is launched. [true/false]
                enabled: true
                # Enter the path to the server's root directory. Continuing from the base directory.
                dir: "hub/"
                # Startup script for the server.
                startup: "java -Xmx512M -jar spigot.jar"
            # The server's name.
            minigames:
                # Enable or disable automatic server startup when the program is launched. [true/false]
                enabled: true
                # Enter the path to the server's root directory. Continuing from the base directory.
                dir: "minigames/"
                # Startup script for the server.
                startup: "java -Xmx512M -jar spigot.jar"
        filters:
            # Enable or disable the custom filters. [true/false]
            enabled: true
        # Options for the backup command.
        backup:
            # Enter the path where the backup files will be placed. This dose not include the base directory.
            dir: "backups/"
    webcon:
        # Enable or disable webcon. [true/false]
        enabled: true
        # Enter the path to webcon's html files.
        dir: "static/web/html/"
        # Which host will webcon run on.
        host: ""
        # Which port will webcon run on.
        port: "80"
        # Add as many users you want here for the webcon login.
        # format: username:password
        users:
            - "admin:changeme"
            - "ThatOneKidEveryoneHates:TheBestPasswordEver"
        blacklist:
            # Blacklist any *WEBCON* users from accessing webcon.
            # NOTE: Do not leave it blank. If you got no users to blacklist, 'users: []' is the way to go. Or else webcon will break.
            users:
                - "ThatOneKidEveryoneHates"
            # Blacklist any IP from accessing webcon.
            # NOTE: Do not leave it blank. If you got no IP to blacklist, 'IP: []' is the way to go. Or else webcon will break.
            IP:
                - "123.45.67.890"
        # Enable or disable any of the messages from webcon. (Best if it's spamming the console too much.) [true/false] for all of them.
        messages:
            login_success: true
            login_fail: true
            ws_connect: true
            ws_disconnect: true
            # Recommended to keep true :)
            exec_command: true

Now that everything's on a config file, customizing GoMinewrap to the way you like it is so much easier.
I have added comments explaining how to use each and every item in the config file. If you still have any questions, you can ask me on the Discord or Skype group chat.
  
# What is the future of GoMinewrap?  
  
Then for later updates, I have a lot of HUGE ideas for GoMinewrap. For example:
  - Multi server type support - Support other server types like Vanilla, SpongePowered, and Bungeecord.
  - In-game color support - Since the program uses stdout from the subprocess, there is no color stuff to make use of :(. But I will find a way sooner or later.
  - Make webcon a proper web based dahsboard - Then anyone and/or everyone can have a web based dashboard for their server :D. But this really won't be easy at all.
  - You tell me, that's all I got :P
  
Alright, there you go, I really hope you like GoMinewrap, I spent a lot of time developing it and not to mention this is my very first GoLang project!  
  
I started development of GoMinewrap just a few days after learning the basics of GoLang and tried out new packages, learned how to use it, the sytaxes and error handling,  
All that while developing GoMinewrap ;D.  
  
I'm sure because of that, the code might be messy or bad or they'll be a lot of bugs or broken things. But don't forget, I straight up jumped to making something huge right after I learned the basics. I slowly learned how to use the http server, subprocesses, regex, http templates and a lot of other amazing GoLang stuff.  
And I know there is a lot of room for improvement ;p.  
  
  
# Update changelogs
  
  **New in v0.2**
  
  No more flags! GoMinewrap now comes with a yaml config file where you can customize anything you want.  
  Webcon now has IP and webcon user blacklisting. You can blacklist IP(s) or users on the config file: *webcon -> blacklist*.
  And of course, webcon has multi user support. You can add as many users you want on the config file: *webcon -> users*.
  
  **New in v0.3**
  
  Multi server support! You can now run as many servers as you want on GoMinewrap.
  
  - Add the servers on the config.yml file.
  - Set the base server directory where all the server folders are.
  - Set the primary server (main server)
  - You're done. If all is done right, GoMinewrap should startup with no errors, all commands relating to managing the servers are listed on the help menu.
  
Webcon also comes with a simple sidebar (side navbar) where you can switch to different servers.
  
# Support me?  
If you like my projects, donating would be highly appreciated!
  - https://www.paypal.me/SuperPykkon
  - SuperPykkon@gmail.com
  
# Stay connected?
Feel free to chat with me, I'm very active on the internet ;o
  - Discord: https://discord.gg/tae9mst
  - Skype: https://join.skype.com/NvciucPmL1lX (why not?)
  - Minecraft IGN: _nullbyte
  
>>>>>>> super/master
