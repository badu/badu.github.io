# A Simple Channels Problem
---

This is experiment is a result of a discussion with a junior, which degenerated in a discussion with chatGPT.

I'm planning to write an article explaining it. The current readme is a TLDR. 

## How to look for comments 
---

Execute the tests and do a comments lookup in this order :

`bcast1_test.go`
`bcast2_noiface_test.go`
`bcast2_pointer_test.go`
`bcast2_iface_test.go`
`bcast2_close_channel_test.go`
`bcast2_iface_heavy_broken_test.go`
`bcast2_iface_heavy_fixed_test.go`
`bcast2_iface_last_test.go`

The conclusions are in `bcast2_iface_last_test.go`.

Enjoy!
