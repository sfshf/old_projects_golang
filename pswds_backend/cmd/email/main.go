package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	req, err := http.NewRequest(http.MethodGet, "https://box.n1xt.net/admin/mail/users?format=json", nil) // https://172.31.87.28:443/; http://172.31.87.28:80/
	if err != nil {
		log.Println(err)
		return
	}
	auth := base64.StdEncoding.EncodeToString([]byte(os.Getenv("PRIVACY_EMAIL_ADMIN_ACCOUNT") + ":" + os.Getenv("PRIVACY_EMAIL_ADMIN_PASSWORD")))
	log.Println("========>Authorization: ", auth)
	req.Header.Set("Authorization", "Basic "+auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("response body: %s\n", respData)
}

func post() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	data := url.Values{}
	data.Add("email", "test007@n1xt.net")
	data.Add("password", "Ro4CtB!W^N9pe7Xr")
	data.Add("privileges", "")
	log.Printf("==========>Request body: %s\n", data.Encode())
	req, err := http.NewRequest(http.MethodPost, "https://box.n1xt.net:993/admin/mail/users/add", strings.NewReader(data.Encode()))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	auth := base64.StdEncoding.EncodeToString([]byte(os.Getenv("PRIVACY_EMAIL_ADMIN_ACCOUNT") + ":" + os.Getenv("PRIVACY_EMAIL_ADMIN_PASSWORD")))
	log.Println("========>Authorization: ", auth)
	req.Header.Set("Authorization", "Basic "+auth)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("response body: %s\n", respData)
}

func fetch1() {
	// data, err := hex.DecodeString("89504E470D0A1A0A")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%s\n", data)
	// return

	cli, err := imapclient.DialTLS("box.n1xt.net:993", nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer cli.Close()
	if err := cli.Login("test1@n1xt.net", "Ro4CtB!W^N9pe7Xr").Wait(); err != nil {
		log.Println(err)
		return
	}

	selectedMbox, err := cli.Select("INBOX", nil).Wait()
	if err != nil {
		log.Printf("failed to select INBOX: %v\n", err)
		return
	}

	if selectedMbox.NumMessages > 0 {
		log.Println("======> selectedMbox.NumMessages:", selectedMbox.NumMessages)
		var seqSet imap.SeqSet
		seqSet.AddNum(selectedMbox.NumMessages)
		fetchOptions := &imap.FetchOptions{
			BodySection: []*imap.FetchItemBodySection{{}},
		}
		fetchCmd := cli.Fetch(seqSet, fetchOptions)
		defer fetchCmd.Close()

		for {
			msg := fetchCmd.Next()
			if msg == nil {
				log.Println("FETCH command did not return any message")
				return
			}
			log.Println("=========>msg.SeqNum:", msg.SeqNum)

			var bodySection imapclient.FetchItemDataBodySection
			ok := false
			for {
				item := msg.Next()
				if item == nil {
					break
				}
				bodySection, ok = item.(imapclient.FetchItemDataBodySection)
				if ok {
					break
				}
			}
			if !ok {
				log.Println("FETCH command did not return body section")
				return
			}

			// Read the message via the go-message library
			mr, err := mail.CreateReader(bodySection.Literal)
			if err != nil {
				log.Printf("failed to create mail reader: %v\n", err)
				return
			}

			// Print a few header fields
			h := mr.Header
			if date, err := h.Date(); err != nil {
				log.Printf("failed to parse Date header field: %v", err)
			} else {
				log.Printf("Date: %v", date)
			}
			if from, err := h.AddressList("From"); err != nil {
				log.Printf("failed to parse To header field: %v", err)
			} else {
				log.Printf("From: %v", from[0])
			}
			subject, err := h.Text("Subject")
			if err != nil {
				log.Printf("failed to parse Subject header field: %v", err)
			} else {
				log.Printf("Subject: %v", subject)
			}

			// Process the message's parts
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					log.Printf("failed to read message part: %v\n", err)
					return
				}

				switch p.Header.(type) {
				case *mail.InlineHeader:
					// This is the message's text (can be plain-text or HTML)
					b, _ := io.ReadAll(p.Body)
					if len(b) > 100 {
						b = b[:100]
					}
					emailTxt := string(b)
					log.Printf("Inline text: %v", emailTxt)
					log.Println("one section")
					// re := regexp.MustCompile(`<h3>([[:digit:]]{6})</h3>`)
					// submatches := re.FindStringSubmatch(emailTxt)
					// log.Printf("submatches: %#v\n", submatches)
				case *mail.AttachmentHeader:
					log.Printf("Attachmenet type: %s", p.Header.(*mail.AttachmentHeader).Get("Content-Type"))
					log.Printf("Attachmenet filename: %s", p.Header.(*mail.AttachmentHeader).Get("Content-Disposition"))
					b, _ := io.ReadAll(p.Body)
					log.Printf("Attachment [%v]: %d",
						p.Header.(*mail.AttachmentHeader).Map(),
						len(b),
					)
				}
			}
		}
	}
	if err := cli.Logout().Wait(); err != nil {
		log.Printf("failed to logout: %v\n", err)
		return
	}
}

func list() {
	cli, err := imapclient.DialTLS("box.n1xt.net:993", nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer cli.Close()
	if err := cli.Login("test1@n1xt.net", "Ro4CtB!W^N9pe7Xr").Wait(); err != nil {
		log.Println(err)
		return
	}

	listCmd := cli.List("", "%", &imap.ListOptions{
		ReturnStatus: &imap.StatusOptions{
			NumMessages: true,
			NumUnseen:   true,
		},
	})
	for {
		mbox := listCmd.Next()
		if mbox == nil {
			break
		}
		log.Printf("Mailbox %q contains %d messages (%d unseen)\n", mbox.Mailbox, *mbox.Status.NumMessages, *mbox.Status.NumUnseen)
	}
	if err := listCmd.Close(); err != nil {
		log.Fatalf("LIST command failed: %v", err)
	}
	if err := cli.Logout().Wait(); err != nil {
		log.Printf("failed to logout: %v\n", err)
		return
	}
}

func fetch2() {
	cli, err := imapclient.DialTLS("box.n1xt.net:993", nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer cli.Close()
	if err := cli.Login("test1@n1xt.net", "Ro4CtB!W^N9pe7Xr").Wait(); err != nil {
		log.Println(err)
		return
	}

	uid := imap.UID(2)
	fetchOptions := &imap.FetchOptions{Envelope: true}

	// Login, select and fetch a message in a single roundtrip
	selectCmd := cli.Select("INBOX", nil)
	fetchCmd := cli.Fetch(imap.UIDSetNum(uid), fetchOptions)
	if _, err := selectCmd.Wait(); err != nil {
		log.Fatalf("failed to select INBOX: %v", err)
	}
	if messages, err := fetchCmd.Collect(); err != nil {
		log.Fatalf("failed to fetch message: %v", err)
	} else {
		for idx, message := range messages {
			log.Printf("Date %d: %v\n", idx, message.Envelope.Date)
			log.Printf("Subject %d: %v\n", idx, message.Envelope.Subject)
			log.Printf("From %d: %v\n", idx, message.Envelope.From)
			log.Printf("Sender %d: %v\n", idx, message.Envelope.Sender)
			log.Printf("ReplyTo %d: %v\n", idx, message.Envelope.ReplyTo)
			log.Printf("To %d: %v\n", idx, message.Envelope.To)
			log.Printf("Cc %d: %v\n", idx, message.Envelope.Cc)
			log.Printf("Bcc %d: %v\n", idx, message.Envelope.Bcc)
			log.Printf("InReplyTo %d: %v\n", idx, message.Envelope.InReplyTo)
			log.Printf("MessageID %d: %v\n", idx, message.Envelope.MessageID)
		}
	}
	if err := cli.Logout().Wait(); err != nil {
		log.Printf("failed to logout: %v\n", err)
		return
	}
}

func expunge() {
	cli, err := imapclient.DialTLS("box.n1xt.net:993", nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer cli.Close()
	if err := cli.Login("test1@n1xt.net", "Ro4CtB!W^N9pe7Xr").Wait(); err != nil {
		log.Println(err)
		return
	}

	// Login, select and fetch a message in a single roundtrip
	selectCmd := cli.Select("INBOX", nil)
	if _, err := selectCmd.Wait(); err != nil {
		log.Fatalf("failed to select INBOX: %v", err)
	}

	seqSet := imap.SeqSetNum(3)
	storeFlags := imap.StoreFlags{
		Op:    imap.StoreFlagsAdd,
		Flags: []imap.Flag{imap.FlagDeleted},
	}
	if err := cli.Store(seqSet, &storeFlags, nil).Close(); err != nil {
		log.Fatalf("Store() = %v", err)
	}

	seqNums, err := cli.Expunge().Collect()
	if err != nil {
		log.Fatalf("Expunge() = %v", err)
	} else if len(seqNums) != 1 || seqNums[0] != 1 {
		log.Printf("Expunge().Collect() = %v, want [1]", seqNums)
	}
	log.Printf("2 Expunge().Collect() = %v, want [1]", seqNums)
	if err := cli.Logout().Wait(); err != nil {
		log.Printf("failed to logout: %v\n", err)
		return
	}
}

func main1() {
	cli, err := imapclient.DialTLS("box.n1xt.net:993", nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer cli.Close()
	if err := cli.Login("test1@n1xt.net", "Ro4CtB!W^N9pe7Xr").Wait(); err != nil {
		log.Println(err)
		return
	}

	type GetPrivacyEmailsResponse_Email struct {
		Uid    string `protobuf:"bytes,1,opt,name=uid,proto3" json:"uid,omitempty"`
		SentBy string `protobuf:"bytes,2,opt,name=sentBy,proto3" json:"sentBy,omitempty"`
		SentAt int64  `protobuf:"varint,3,opt,name=sentAt,proto3" json:"sentAt,omitempty"`
		Title  string `protobuf:"bytes,4,opt,name=title,proto3" json:"title,omitempty"`
	}

	var total int64
	mailboxes := []string{"INBOX", "Spam", "Trash", "Archive"}
	listResult, err := cli.List("", "%", &imap.ListOptions{
		ReturnStatus: &imap.StatusOptions{
			NumMessages: true,
		},
	}).Collect()
	if err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println(err)
			return
		}
	}
	for _, item := range listResult {
		if slices.ContainsFunc(mailboxes, func(mailbox string) bool {
			return strings.EqualFold(strings.TrimSpace(item.Mailbox), mailbox)
		}) {
			log.Println("============>*item.Status.NumMessages:", *item.Status.NumMessages)
			total += int64(*item.Status.NumMessages)
		}
	}
	log.Println("================>total:", total)
	var pageNum int64
	var pageSize int64 = 20
	start := pageNum*pageSize + 1
	if start > total {
		log.Println("page error: start > total")
		return
	}
	stop := (pageNum+1)*pageSize + 1
	if stop > total {
		stop = total
	}
	var count int64
	var list []*GetPrivacyEmailsResponse_Email
	for _, mailbox := range mailboxes {
		selectedMbox, err := cli.Select(mailbox, nil).Wait()
		if err != nil {
			log.Println(err)
			return
		}
		if selectedMbox.NumMessages > 0 {
			for i := selectedMbox.NumMessages; i > 0; i-- {
				count++
				if count >= start && count < stop {
					var one GetPrivacyEmailsResponse_Email
					seqSet := imap.SeqSetNum(i)
					fetchOptions := &imap.FetchOptions{
						BodySection: []*imap.FetchItemBodySection{{}},
					}
					fetchCmd := cli.Fetch(seqSet, fetchOptions)
					for {
						msg := fetchCmd.Next()
						if msg == nil {
							break
						}
						one.Uid = fmt.Sprintf("%s-%d", mailbox, msg.SeqNum) // uid
						var bodySection imapclient.FetchItemDataBodySection
						ok := false
						for {
							item := msg.Next()
							if item == nil {
								break
							}
							bodySection, ok = item.(imapclient.FetchItemDataBodySection)
							if ok {
								break
							}
						}
						if !ok {
							break
						}
						mr, err := mail.CreateReader(bodySection.Literal)
						if err != nil {
							log.Println(err)
							return
						}
						h := mr.Header
						// sentAt
						if date, err := h.Date(); err != nil {
							log.Println(err)
							return
						} else {
							one.SentAt = date.UnixMilli()
						}
						// sentBy
						if from, err := h.AddressList("From"); err != nil {
							log.Println(err)
							return
						} else {
							if len(from) > 0 {
								var sb strings.Builder
								for idx, item := range from {
									sb.WriteString(item.Name)
									sb.WriteRune(' ')
									sb.WriteString(item.Address)
									if idx != len(from)-1 {
										sb.WriteRune('\n')
									}
								}
								one.SentBy = sb.String()
							}
						}
						// title
						subject, err := h.Text("Subject")
						if err != nil {
							log.Println(err)
							return
						} else {
							one.Title = subject
						}
					}
					if err := fetchCmd.Close(); err != nil {
						log.Println(err)
						return
					}
					list = append(list, &one)
				} else {
					break
				}
			}
		}
	}
	for _, item := range list {
		log.Printf("SentAt %v \tSentBy %v\tTitle %v\tUid %v\t\n", item.SentAt, item.SentBy, item.Title, item.Uid)
	}

}
