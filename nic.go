package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/viper"
)

const NIC_CL_NEW_DOMAINS = "https://www.nic.cl/registry/Ultimos.do?t=1h&f=csv"
const NIC_CL_DELETED_DOMAINS = "https://www.nic.cl/registry/Eliminados.do?t=1d&f=txt"
const NIC_CL_ARBITRAJE = "https://www.nic.cl/rcal/sentenciasArbitrales.do?arbitro=&consignado=1&dominio=&filtroLimit=20&page=0&totalFallos=5"

const MAX_SENTENCE_NAMES_LENGTH = 150

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
🆕 [Nuevo Dominio]

Se acaba de registrar el dominio %s.

Más información acá: https://www.nic.cl/registry/Whois.do?d=%s
				`,
				newDomain[0],
				url.PathEscape(newDomain[0]),
			)
			for _, poster := range posters {
				err := poster.Post(message)
				if err != nil {
					log.Printf("error posting: %s", err)
				}
			}
			viper.Set("general.lastCreatedDate", newDomain[1])
			time.Sleep(1 * time.Second)
		}
	} else {
		log.Printf("no new domains since last minute :(")
	}
	viper.WriteConfig()
	return nil
}

func monitorArbitraje(posters []Poster) error {
	latest := viper.GetString("general.lastarbitraje")
	resp, err := http.Get(NIC_CL_ARBITRAJE)
	if err != nil {
		return fmt.Errorf("cannot get arbitraje: %s", err)
	}
	jsonBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot get arbitraje: %s", err)
	}
	var arb Arbitraje
	err = json.Unmarshal(jsonBytes, &arb)
	if err != nil {
		return fmt.Errorf("cannot get arbitraje: %s", err)
	}
	newFallos := make([]*Fallo, 0)

	for _, fallo := range arb.Fallos {
		if fallo.NombreDominio == latest {
			break
		}
		fallo.AdjustNames()
		newFallos = append(newFallos, fallo)
	}

	if len(newFallos) > 0 {
		for i := len(newFallos) - 1; i >= 0; i-- {
			fallo := newFallos[i]
			message := fmt.Sprintf(
				`
⚔️ [Controversia Resuelta]

Por dominio %s entre:
`,
				fallo.NombreDominio,
			)
			for i, parte := range fallo.LstPartes {
				emoji := "❌"
				if (i + 1) == fallo.GanaRevocante {
					emoji = "✅"
				}
				message += fmt.Sprintf("%s %s\n", emoji, parte.Nombre)
			}
			message += fmt.Sprintf("\nSentencia disponible en https://www.nic.cl/rcal/downloadResolucion.do?uuid=%s", fallo.ArchivoSentencia)
			for _, poster := range posters {
				err := poster.Post(message)
				if err != nil {
					log.Printf("error posting: %s", err)
				}
			}
			viper.Set("general.lastarbitraje", fallo.NombreDominio)
			time.Sleep(1 * time.Second)
		}
	} else {
		log.Printf("no new sentencias since last minute :(")
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
		// Post equispaced since now until midnight
		now := time.Now()
		domainsNumber := len(newDomains)
		rand.Seed(now.UnixNano())
		rand.Shuffle(domainsNumber, func(i, j int) { newDomains[i], newDomains[j] = newDomains[j], newDomains[i] })
		untilMidnight := time.Until(time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local))
		period := time.Duration(int(untilMidnight) / domainsNumber)
		log.Printf("%d domains and %d until midnight... Must wait %d seconds between domains", domainsNumber, untilMidnight, period)
		for _, newDomain := range newDomains {
			message := fmt.Sprintf(
				`
❌ [Dominio Eliminado]

Hoy se liberó el dominio %s.

Puedes registrarlo acá: https://www.nic.cl/registry/Whois.do?d=%s&buscar=Submit+Query&a=inscribir
				`,
				newDomain,
				url.PathEscape(newDomain),
			)
			for _, poster := range posters {
				err := poster.Post(message)
				if err != nil {
					log.Printf("error posting: %s", err)
				}
			}
			log.Printf("sleeping %d seconds...", period)
			time.Sleep(period)
		}
	}
	return nil
}
