package resolvconf

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	reGoodChars = regexp.MustCompile(`^[a-zA-Z0-9\.-]+$`)
	reGoodNums  = regexp.MustCompile(`^[0-9\.]+$`)
)

func Run(_ context.Context, l *log.Logger) error {
	var searchDomain string
	var nameservers []string

	for _, e := range os.Environ() {
		kv := strings.SplitN(e, "=", 2)
		if len(kv) != 2 {
			continue
		}

		if kv[0] == "PSA_DHCPC_DOMAIN_NAME" && reGoodChars.MatchString(kv[1]) {
			searchDomain = kv[1]
		}
		if kv[0] == "PSA_DHCPC_DNS_LIST" && len(kv[1]) > 0 {
			for _, ns := range strings.Split(kv[1], ",") {
				if reGoodNums.MatchString(ns) {
					nameservers = append(nameservers, ns)
				}
			}
		}
	}

	// only update if we do have an NS list.
	if len(nameservers) == 0 {
		return nil
	}

	buf := []byte("# written by psa-dhcpc\n")
	if len(searchDomain) > 0 {
		buf = append(buf, []byte(fmt.Sprintf("search %s\n", searchDomain))...)
	}
	for _, ns := range nameservers {
		buf = append(buf, []byte(fmt.Sprintf("nameserver %s\n", ns))...)
	}
	return update(buf)
}

func update(buf []byte) (err error) {
	tmpfh, err := ioutil.TempFile("/etc", "resolvconf-*.tmp")
	if err != nil {
		return err
	}

	name := tmpfh.Name()
	defer func() {
		if err != nil {
			os.Remove(name)
		}
	}()

	nr, werr := tmpfh.Write(buf)
	cerr := tmpfh.Close()
	if werr != nil {
		err = werr
		return
	}
	if cerr != nil {
		err = cerr
		return
	}
	if nr != len(buf) {
		err = io.ErrShortWrite
		return
	}
	if err = os.Chmod(name, 0644); err != nil {
		return
	}
	if err = os.Rename(name, "/etc/resolv.conf"); err != nil {
		return
	}
	return nil
}
