package factcheck

import (
	"fmt"
	domainutil "github.com/bobesa/go-domain-util/domainutil"
	"github.com/bwmarrin/discordgo"
	"github.com/ethanent/discordkvs"
	xurls "mvdan.cc/xurls/v2"
	"net/url"
	"strings"
)

var strictURLRegex = xurls.Strict()

func MessageSend(s *discordgo.Session, e *discordgo.MessageCreate, app *discordkvs.Application) {
	urls := strictURLRegex.FindAllString(e.Content, 2)

	if len(urls) == 0 {
		return
	}

	// Check if fact check is enabled in guild

	fcv, err := app.Get(e.GuildID, "factcheck")

	if err != nil {
		return
	}

	if len(fcv) < 1 || fcv[0] != 1 {
		// Fact check not enabled

		return
	}

	// Perform fact check for each URL

	for _, u := range urls {
		curUrl, err := url.Parse(u)

		if err != nil {
			continue
		}

		// Ensure that URL path is long enough to qualify as a potential article

		if len(curUrl.Path) < 6 {
			continue
		}

		// Filter host

		curUrl.Host = domainutil.Domain(curUrl.Hostname())

		// Locate sites within host

		hostSites, ok := fcdata[curUrl.Hostname()]

		if !ok {
			// Not a fact-checked site

			continue
		}

		// Find best-matching site

		var match *SiteData
		bestMatchLen := -1

		for _, s := range hostSites {
			if strings.Index(curUrl.Path, s.URL.Path) == 0 {
				curMatchLen := len(s.URL.Path)

				if bestMatchLen == -1 || curMatchLen > bestMatchLen {
					match = s
					bestMatchLen = curMatchLen
				}
			}
		}

		if match == nil {
			continue
		}

		fmt.Println(match.DisplayName, match.Accuracy, match.Bias)

		// Strong sources

		// 0 = unclassified, 1 = top tier, 2 = great, 3 = fine, 4 = poor, 5 = very poor, 6 = subpar
		sourceQuality := 0

		if match.Bias == "pro-science" {
			if match.Accuracy == "very high" {
				sourceQuality = 1
			} else {
				sourceQuality = 2
			}
		}

		if match.Bias == "least biased" || match.Bias == "left-center" || match.Bias == "right-center" {
			if match.Accuracy == "very high" {
				sourceQuality = 2
			} else if match.Accuracy == "high" {
				sourceQuality = 3
			}
		}

		// Weak sources

		if match.Bias == "questionable" || match.Bias == "conspiracy/pseudoscience" {
			sourceQuality = 5
		}

		if match.Accuracy == "low" || match.Accuracy == "very low" {
			sourceQuality = 4
		}

		if match.Accuracy == "mixed" {
			if match.Bias == "left" || match.Bias == "right" {
				sourceQuality = 4
			} else {
				sourceQuality = 6
			}
		}

		switch sourceQuality {
		case 1:
			s.MessageReactionAdd(e.ChannelID, e.Message.ID, "✅")
			s.MessageReactionAdd(e.ChannelID, e.Message.ID, "🧠")
		case 2:
			s.MessageReactionAdd(e.ChannelID, e.Message.ID, "✅")
		case 3:
			s.MessageReactionAdd(e.ChannelID, e.Message.ID, "☑️")
		case 6:
			s.MessageReactionAdd(e.ChannelID, e.Message.ID, "🤐")
		case 4:
			fallthrough
		case 5:
			sendEmbed := &discordgo.MessageEmbed{
				Title:       ":warning: Poor Source Alert: " + match.DisplayName,
				Description: "It is recommended that you attempt to locate this information or news from another source.",
				Fields: []*discordgo.MessageEmbedField{
					&discordgo.MessageEmbedField{
						Name:   "Accuracy",
						Value:  match.Accuracy,
						Inline: true,
					},
					&discordgo.MessageEmbedField{
						Name:   "Bias",
						Value:  match.Bias,
						Inline: true,
					},
				},
				Color: 0xC20000,
				Footer: &discordgo.MessageEmbedFooter{
					Text:         "Media Bias Fact Check, with alterations",
				},
			}

			if sourceQuality == 5 {
				sendEmbed.Color = 0x9C0000
				sendEmbed.Title = ":warning::exclamation: Extremely Poor Source Alert: " + match.DisplayName
			}

			s.MessageReactionAdd(e.ChannelID, e.Message.ID, "⚠️")

			s.ChannelMessageSendComplex(e.ChannelID, &discordgo.MessageSend{
				Embed: sendEmbed,
			})
		}
	}
}
