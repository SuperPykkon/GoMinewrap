# What is GoMinewrap?
GoMinewarp is a wrapper for your Minecraft server console. It comes with many new features that could possibly make server management easier.  
GoMinewarp is also a re-creation of 2 of my very old projects;  
  - minecraft-server-web-console: https://github.com/SuperPykkon/minecraft-server-web-console  
  - minecraft-server-wrapper: https://github.com/SuperPykkon/minecraft-server-wrapper  

**What does it do?**  
It completly changes the console log format and adds icons, highlighting errors, warnings, player join/leave and much more.  
It also comes with a fully functional real time web based console which also has icons, highlighting errors, warnings, player join/leave etc etc.  
The web based console, also known as *webcon* can be toggled on or off with a flag *--webcon  [on/off]*.  
It also has an authentication system where the username and password can easily be set with the flags *--webconUser [user]* and *--webconPass [password]*.  

# Screenshots
Here is how it looks like: 
  - http://prntscr.com/ef33y0 - webcon login, invalid username/password.
  - http://prntscr.com/ef346k - how webcon and the main console looks like.
  - http://prntscr.com/ef35ys - how errors look like (done on purpose).
  - http://prntscr.com/ef38e7 - how player join/leave and chat messages look like. Including command highlighting. Very colourful :D

# How to use it?  
Firsly, you have the option to use the .go file or the executables. Currently, I've made executables for Windows and Linux which you can download here:  
  - Windows: https://github.com/SuperPykkon/GoMinewrap/releases/download/v0.1/mcs-v0.1-windows.exe  
  - Linux: https://github.com/SuperPykkon/GoMinewrap/releases/download/v0.1/mcs-v0.1-linux  
  - Mac: https://github.com/SuperPykkon/GoMinewrap/releases/download/v0.1/mcs-v0.1-mac  
  
  **NOTE**
  Please note that the executables for Linux and Mac is **not tested**. I don't have any of them and I can't get a VM either.  
  So, it is very much possible that it won't work. I recommend you to run mcs.go instead or just use Wine on the Windows executable.  
  A proper working version of these will be released soon!  
  
Then, if you want to use the web based console, download the folder "static" with all the html files for webcon and place it in the same directory.
You can also place the files else where, but you have to make sure to provide the new path with the flag *--webconDir path/to/webconFiles/*.  
  
You will now have to setup the Minecraft server. You can place the executable anywhere, just make sure to set the path/to/serverFiles/ with the flag *--ServerDir /path/to/server*.  
If you're placing the execuable in the same directory as spigot.jar, use *--serverDir .*  
By default, the startup script is **java -Xmx1G -jar spigot.jar** but you can change it with the flag *--runCmd [startup script]*.

# All the flags
    1. --runCmd [startup script]
       Startup script for the Minecraft server.
       default: java -Xmx1G -jar spigot.jar
    
    2. --serverDir [path/to/serverFiles/]
       Set the directory of the server.
       default: server/
       
    3. --webcon [on/off]
       Enable or disable webcon (web console).
       default: off
    
    4. --webconDir [path/to/webconFiles/]
       Set the directory of the webcon files.
       default: static/web/html/
    
    5. --webconHost
       Set the host for webcon web server.
       default: "" (nothing)
       
    6. --webconPort
       Set the port for webcon web server.
       default: 80
       
    7. --webconUser
       Add a username for webcon login.
       default: admin
       
    8. --webconPass
       Set a password for the webcon account.
       default: changeme

So, there are a lot of flags provided for more control, I have many many more planned for future updates.  
  
# What is the future of GoMinewarp?  

I have a LOT of amazing ideas for GoMinewarp. Firsly, what I plan for the next update;  
  - To figure out how to use Cobra and add a config file for GoMinewarp. This will make things way easier, you won't have to worry about a bunch of flags to start the server.  ure out how to use Cobra and add a config file for GoMinewarp. This will make things way easier, you won't have to worry about a bunch of flags to start the server.  
  - Multi user account support for webcon - once I get the config file working, I will then add the option to add multiple accounts for webcon login on the config file.
  - Bug fixes? I'm sure there will be a lot of bugs :P
  
Then for later updates, I have a lot of HUGE ideas for GoMinewarp. For example:
  - Multi server type support - Support other server types like Vanilla, SpongePowered, and Bungeecord.
  - Multi server support - Support running multiple servers at once, switch servers using !server [server], execute a command on all the servers, and much much more!
  - In-game color support - Since the program uses stdout from the subprocess, there is no color stuff to make use of :(. But I will find a way sooner or later.
  - You tell me, that's all I got :P
  
Alright, there you go, I really hope you like GoMinewarp, I spent a lot of time developing it and not to mention this is my very first GoLang project!  
I started development of GoMinewarp just a few days after learning the basics of GoLang and tried out new packages, learned how to use it, the sytaxes and error handling,  
All that while developing GoMinewarp ;D.  
  
I'm sure because of that, the code might be messy or bad or they'll be a lot of bugs or broken things. But don't forget, I straight up jumped to making
something huge right after I learned the basics. I slowly learned how to use the http server, subprocesses, regex, http templates and a lot of other amazing GoLang stuff.  
And I know there is a lot of room for improvement ;p.  

# Stay connected?
Feel free to chat with me, I'm very active on the internet ;o
  - Discord: https://discord.gg/tae9mst
  - Skype: https://join.skype.com/NvciucPmL1lX (why not?)
  - Minecraft IGN: _nullbyte
  
