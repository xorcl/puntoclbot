package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

const NIC_CL_NEW_DOMAINS = "https://www.nic.cl/registry/Ultimos.do?t=1h&f=csv"
const NIC_CL_DELETED_DOMAINS = "https://www.nic.cl/registry/Eliminados.do?t=1d&f=txt"

func monitorNewDomains(posters []Poster) error {
	latest := viper.GetString("general.lastCreatedDate")
	resp, err := http.Get(NIC_CL_NEW_DOMAINS)
	if err != nil {
		return fmt.Errorf("cannot get new domains: %s", err)
	}
	lines := csv.NewReader(resp.Body)
	newDomains := make([][]string, 0)
	// Skip first line
	lines.Read()
	for {
		line, err := lines.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Printf("error reading new domain line :%s", err)
				continue
			}
		}
		if len(line) != 2 {
			log.Printf("error reading new domain line: wrong length: %+v", line)
			continue
		}
		if line[1] > latest {
			newDomains = append(newDomains, line)
		} else {
			break
		}
	}
	if len(newDomains) > 0 {
		for i := len(newDomains) - 1; i >= 0; i-- {
			newDomain := newDomains[i]
			message := fmt.Sprintf(
				`
Alguien acaba de registrar el dominio %s.

Más información acá: https://www.nic.cl/registry/Whois.do?d=%s
				`,
				newDomain[0],
				newDomain[0],
			)
			for _, poster := range posters {
				err := poster.Post(message)
				if err != nil {
					log.Printf("error posting: %s", err)
				}
			}
			viper.Set("general.lastCreatedDate", newDomain[1])
		}
	} else {
		log.Printf("no new domains since last minute :(")
	}
	viper.WriteConfig()
	return nil
}

func monitorDeletedDomains(posters []Poster) error {
	resp, err := http.Get(NIC_CL_DELETED_DOMAINS)
	if err != nil {
		return fmt.Errorf("cannot get deleted domains: %s", err)
	}
	lines := csv.NewReader(resp.Body)
	newDomains := make([]string, 0)
	// Skip first line
	lines.Read()
	for {
		line, err := lines.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Printf("error reading new domain line :%s", err)
				continue
			}
		}
		if len(line) != 1 {
			log.Printf("error reading new domain line: wrong length: %+v", line)
			continue
		}
		newDomains = append(newDomains, line[0])
	}
	if len(newDomains) > 0 {
		for _, newDomain := range newDomains {
			message := fmt.Sprintf(
				`
Se acaba de liberar el dominio %s.

Puedes registrarlo acá: https://www.nic.cl/registry/Whois.do?d=%s&buscar=Submit+Query&a=inscribir
				`,
				newDomain,
				newDomain,
			)
			// Post equispaced since now until midnight
			now := time.Now()
			domainsNumber := len(newDomain)
			untilMidnight := time.Until(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local))
			period := time.Duration(int(untilMidnight) / domainsNumber)
			log.Printf("%d domains and %d until midnight... Must wait %d seconds between domains", domainsNumber, untilMidnight, period)
			for _, poster := range posters {
				err := poster.Post(message)
				if err != nil {
					log.Printf("error posting: %s", err)
				}
				log.Printf("sleeping %d seconds...", period)
				time.Sleep(period)
			}
		}
	}
	return nil
}
