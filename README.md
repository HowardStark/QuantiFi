# QuantiFi
Tracks total network data usage to help avoid overage fees on data-limited internet/cable plans. QuantiFi is designed to be ran on a Raspberry Pi with two wifi cards, giving it the power to fully monitor data usage. It is unique from all other data usage tracking solutions not already built into your router, as QuantiFi can measure the data-usage accross the entire network (with the notable exception of networks with network switches). 

## How it works
The Raspberry Pi uses one wifi card to actively monitor and measure all incoming network data, while the other hosts the local server from which you can view your current data useage. The monitoring is accomplished by putting the wifi card into promiscuous mode, allowing the network interface to passively monitor all network traffic both inbound and outbound. To avoid counting outbound packets against your data, the system keeps a running list of the MAC addresses of the devices on the network, and checks to ensure that the sender was not on the network and the receiver was. 

All that is the "abstract" implementation details. The way I accomplished this was by using small wrappers around Linux networking tools. This includes interacting with and controlling the network interfaces on a lower level, turning them off and ensuring that they connect to the proper network as well as actually capturing the data (The concepts I use here have since been transformed into a far more comprehensive cross-platform network interface manager, (WifiManager)[https://github.com/ottopress/WifiManage]). 

One of the issues I ran into while developing this was trying to determine which network interface was actually internet connected. This proved to be far more difficult than I expected, what with naming inconsistencies and other nonsense. Though the actual implementation varied ((network.go:138)[https://github.com/HowardStark/QuantiFi/blob/a1f9635de097d68d7a3f08ec7dff4ef94dc64b52/network.go#L138]), the idea was to use the platform specific `route` command, and then scrape the command output for the interface. While a hack, no doubt, this ended up working surprisingly well.

For the server portion, it simply serves a frontend which is powered by a JSON-api. This has the added benefit of allowing people to hook into the system and trigger, say, events as they please. Pretty straightforward and simple.

## Important

This tool is not 100% accurate. The tool will miss some data here, or track more data out of each packet than your ISP. Regrettably, there isn't a lot of insight into their metrics for 'data', so this tool is more of a ballpark. 
