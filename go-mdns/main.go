package main

import (
	"github.com/alx696/go-mdns/im"
	"log"
)

func main() {
	log.Println("在内网用mDNS发现节点并通信")

	//txt := dns.DigShort("iim.app.lilu.red", 16)
	//log.Println(txt)

	// /sdcard/android/data/red.lilu.red.iim/cache
	go im.Init("./config", "电脑", "data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiIHN0YW5kYWxvbmU9Im5vIj8+CjxzdmcKICAgeG1sbnM6ZGM9Imh0dHA6Ly9wdXJsLm9yZy9kYy9lbGVtZW50cy8xLjEvIgogICB4bWxuczpjYz0iaHR0cDovL2NyZWF0aXZlY29tbW9ucy5vcmcvbnMjIgogICB4bWxuczpyZGY9Imh0dHA6Ly93d3cudzMub3JnLzE5OTkvMDIvMjItcmRmLXN5bnRheC1ucyMiCiAgIHhtbG5zOnN2Zz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciCiAgIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIKICAgeG1sbnM6c29kaXBvZGk9Imh0dHA6Ly9zb2RpcG9kaS5zb3VyY2Vmb3JnZS5uZXQvRFREL3NvZGlwb2RpLTAuZHRkIgogICB4bWxuczppbmtzY2FwZT0iaHR0cDovL3d3dy5pbmtzY2FwZS5vcmcvbmFtZXNwYWNlcy9pbmtzY2FwZSIKICAgaWQ9IkNhcGFfMSIKICAgZW5hYmxlLWJhY2tncm91bmQ9Im5ldyAwIDAgNTEyIDUxMiIKICAgaGVpZ2h0PSI1MTIiCiAgIHZpZXdCb3g9IjAgMCA1MTIgNTEyIgogICB3aWR0aD0iNTEyIgogICB2ZXJzaW9uPSIxLjEiCiAgIHNvZGlwb2RpOmRvY25hbWU9IndhcmVob3VzZS5zdmciCiAgIGlua3NjYXBlOnZlcnNpb249IjAuOTIuNSAoMC45Mi41KzY4KSI+CiAgPG1ldGFkYXRhCiAgICAgaWQ9Im1ldGFkYXRhMTEiPgogICAgPHJkZjpSREY+CiAgICAgIDxjYzpXb3JrCiAgICAgICAgIHJkZjphYm91dD0iIj4KICAgICAgICA8ZGM6Zm9ybWF0PmltYWdlL3N2Zyt4bWw8L2RjOmZvcm1hdD4KICAgICAgICA8ZGM6dHlwZQogICAgICAgICAgIHJkZjpyZXNvdXJjZT0iaHR0cDovL3B1cmwub3JnL2RjL2RjbWl0eXBlL1N0aWxsSW1hZ2UiIC8+CiAgICAgIDwvY2M6V29yaz4KICAgIDwvcmRmOlJERj4KICA8L21ldGFkYXRhPgogIDxkZWZzCiAgICAgaWQ9ImRlZnM5IiAvPgogIDxzb2RpcG9kaTpuYW1lZHZpZXcKICAgICBwYWdlY29sb3I9IiNmZmZmZmYiCiAgICAgYm9yZGVyY29sb3I9IiM2NjY2NjYiCiAgICAgYm9yZGVyb3BhY2l0eT0iMSIKICAgICBvYmplY3R0b2xlcmFuY2U9IjEwIgogICAgIGdyaWR0b2xlcmFuY2U9IjEwIgogICAgIGd1aWRldG9sZXJhbmNlPSIxMCIKICAgICBpbmtzY2FwZTpwYWdlb3BhY2l0eT0iMCIKICAgICBpbmtzY2FwZTpwYWdlc2hhZG93PSIyIgogICAgIGlua3NjYXBlOndpbmRvdy13aWR0aD0iMjQ5MyIKICAgICBpbmtzY2FwZTp3aW5kb3ctaGVpZ2h0PSIxMzg1IgogICAgIGlkPSJuYW1lZHZpZXc3IgogICAgIHNob3dncmlkPSJmYWxzZSIKICAgICBpbmtzY2FwZTp6b29tPSIwLjQ2MDkzNzUiCiAgICAgaW5rc2NhcGU6Y3g9Ii0zNi44ODEzNTYiCiAgICAgaW5rc2NhcGU6Y3k9IjI1NiIKICAgICBpbmtzY2FwZTp3aW5kb3cteD0iMCIKICAgICBpbmtzY2FwZTp3aW5kb3cteT0iMjciCiAgICAgaW5rc2NhcGU6d2luZG93LW1heGltaXplZD0iMSIKICAgICBpbmtzY2FwZTpjdXJyZW50LWxheWVyPSJDYXBhXzEiIC8+CiAgPGcKICAgICBpZD0iZzQiCiAgICAgdHJhbnNmb3JtPSJtYXRyaXgoMC44NzkyNjcwOSwwLDAsMC44NzkyNjcwOSwzMC45MDgwNjUsMzAuOTA3NjI1KSI+CiAgICA8cGF0aAogICAgICAgZD0iTSA0NzQuNjg0LDAgSCA0NTAuNTYgYyAtOC4yMjYsMCAtMTQuOTE4LDYuNjkyIC0xNC45MTgsMTQuOTE4IFYgMzgxLjk2NiBIIDQwOC42NzQgViAyMzcuMzU5IGMgMCwtNC4xNDIgLTMuMzU3LC03LjUgLTcuNSwtNy41IC00LjE0MiwwIC03LjUsMy4zNTggLTcuNSw3LjUgdiA2LjE1MiBIIDI2My41IFYgMTIwLjA1NiBoIDM2Ljg3IHYgMzMuOTMzIGMgMCw0LjE0MiAzLjM1Nyw3LjUgNy41LDcuNSBoIDQxLjQzNCBjIDQuMTQzLDAgNy41LC0zLjM1OCA3LjUsLTcuNSB2IC0zMy45MzMgaCAzMy4zNzMgYyAxLjkyOSwwIDMuNDk3LDEuNTY5IDMuNDk3LDMuNDk4IHYgODMuODA2IGMgMCw0LjE0MiAzLjM1OCw3LjUgNy41LDcuNSA0LjE0MywwIDcuNSwtMy4zNTggNy41LC03LjUgdiAtODMuODA2IGMgMCwtMTAuMiAtOC4yOTgsLTE4LjQ5OCAtMTguNDk3LC0xOC40OTggSCAxMjEuODIzIGMgLTEwLjE5OSwwIC0xOC40OTcsOC4yOTggLTE4LjQ5NywxOC40OTggViAzODEuOTY3IEggNzYuMzU4IFYgMTQuOTE4IEMgNzYuMzU4LDYuNjkyIDY5LjY2NiwwIDYxLjQ0LDAgSCAzNy4zMTYgQyAyOS4wOSwwIDIyLjM5OCw2LjY5MiAyMi4zOTgsMTQuOTE4IHYgOTAuNDE0IGMgMCw0LjE0MiAzLjM1OCw3LjUgNy41LDcuNSA0LjE0MywwIDcuNSwtMy4zNTggNy41LC03LjUgViAxNSBoIDIzLjk2IFYgNDk3IEggMzcuMzk4IFYgMTM1LjMzMiBjIDAsLTQuMTQyIC0zLjM1NywtNy41IC03LjUsLTcuNSAtNC4xNDIsMCAtNy41LDMuMzU4IC03LjUsNy41IHYgMzYxLjc1IGMgMCw4LjIyNiA2LjY5MiwxNC45MTggMTQuOTE4LDE0LjkxOCBIIDYxLjQ0IGMgOC4yMjYsMCAxNC45MTgsLTYuNjkyIDE0LjkxOCwtMTQuOTE4IHYgLTU4LjI3NSBoIDM1OS4yODMgdiA1OC4yNzUgYyAwLDguMjI2IDYuNjkyLDE0LjkxOCAxNC45MTgsMTQuOTE4IGggMjQuMTI0IGMgOC4yMjYsMCAxNC45MTgsLTYuNjkyIDE0LjkxOCwtMTQuOTE4IFYgMTQuOTE4IEMgNDg5LjYwMiw2LjY5MiA0ODIuOTA5LDAgNDc0LjY4NCwwIFogTSAzMTUuMzcsMjU4LjUxMSBoIDI2LjQzNCB2IDI2LjQzMyBIIDMxNS4zNyBaIG0gLTE1LDAgdiAzMy45MzMgYyAwLDQuMTQyIDMuMzU3LDcuNSA3LjUsNy41IGggNDEuNDM0IGMgNC4xNDMsMCA3LjUsLTMuMzU4IDcuNSwtNy41IHYgLTMzLjkzMyBoIDM2Ljg3IFYgMzgxLjk2NiBIIDI2My41IFYgMjU4LjUxMSBaIE0gMzQxLjgwNCwxNDYuNDg5IEggMzE1LjM3IHYgLTI2LjQzMyBoIDI2LjQzNCB6IE0gMTcwLjE5NiwxMjAuMDU2IGggMjYuNDM0IHYgMjYuNDMzIGggLTI2LjQzNCB6IG0gLTUxLjg3LDMuNDk4IGMgMCwtMS45MjkgMS41NjgsLTMuNDk4IDMuNDk3LC0zLjQ5OCBoIDMzLjM3MyB2IDMzLjkzMyBjIDAsNC4xNDIgMy4zNTcsNy41IDcuNSw3LjUgaCA0MS40MzQgYyA0LjE0MywwIDcuNSwtMy4zNTggNy41LC03LjUgViAxMjAuMDU2IEggMjQ4LjUgViAyNDMuNTExIEggMTE4LjMyNiBaIG0gNTEuODcsMTM0Ljk1NyBoIDI2LjQzNCB2IDI2LjQzMyBoIC0yNi40MzQgeiBtIC01MS44NywwIGggMzYuODcgdiAzMy45MzMgYyAwLDQuMTQyIDMuMzU3LDcuNSA3LjUsNy41IGggNDEuNDM0IGMgNC4xNDMsMCA3LjUsLTMuMzU4IDcuNSwtNy41IFYgMjU4LjUxMSBIIDI0OC41IFYgMzgxLjk2NiBIIDExOC4zMjYgWiBNIDc2LjM1OCw0MjMuODA3IHYgLTI2Ljg0MSBoIDM1OS4yODMgdiAyNi44NDEgeiBNIDQ3NC42MDIsNDk3IGggLTIzLjk2IFYgMTUgaCAyMy45NiB6IgogICAgICAgaWQ9InBhdGgyIgogICAgICAgaW5rc2NhcGU6Y29ubmVjdG9yLWN1cnZhdHVyZT0iMCIgLz4KICA8L2c+Cjwvc3ZnPgo=")

	//go func() {
	//	time.Sleep(time.Second * 6)
	//	im.Destroy()
	//}()

	//go func() {
	//	for {
	//		msg := im.ExtractMessage()
	//		log.Println("提取消息:", msg)
	//
	//		time.Sleep(time.Second * 3)
	//	}
	//}()

	//go func() {
	//	for {
	//		ps := im.FindPeer()
	//		log.Println("现有节点:", ps)
	//
	//		var ids []string
	//		err := json.Unmarshal([]byte(ps), &ids)
	//		if err != nil {
	//			log.Println(err)
	//			continue
	//		}
	//		for _, v := range ids {
	//			//err := im.SendText(v, "你好")
	//			//err := im.SendFile(v, "/home/km/下载/s.txt")
	//			infoStr, err := im.GetInfo(v)
	//			if err != nil {
	//				log.Println(err)
	//				continue
	//			}
	//			log.Println("信息:", infoStr)
	//		}
	//
	//		time.Sleep(time.Second * 6)
	//	}
	//}()

	select {}
}
