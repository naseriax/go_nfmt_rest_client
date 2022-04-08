package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type NeSwList struct {
	Items []NEs `json:"items"`
}

type NEs struct {
	NeType             string `json:"neType"`
	NeLabel            string `json:"neLabel"`
	SecondaryCurStatus string `json:"secondaryCurStatus"`
	PrimaryCurStatus   string `json:"primaryCurStatus"`
	SecondarySWVersion string `json:"secondarySWVersion"`
	PrimarySWVersion   string `json:"primarySWVersion"`
}

func getNEs(agent RestAgent) NeSwList {
	var allNEsJson NeSwList
	allNEsRaw := agent.Get("/data/swim/neSoftware", map[string]string{"Range": "items=0-1999"})
	json.Unmarshal([]byte(allNEsRaw), &allNEsJson)

	return allNEsJson
}

func neList(allNEsJson NeSwList) []NEs {
	var neList []NEs
	for _, ne := range allNEsJson.Items {
		if ne.NeType == "1830PSS-PHN" {
			neList = append(neList, ne)
		}
	}
	return neList
}

func exportPrep(allNEsJson NeSwList) [][]string {
	output := [][]string{
		{
			"No",
			"NE Name",
			"Software Version",
		},
	}
	c := 1
	for _, ne := range allNEsJson.Items {
		neName := ne.NeLabel
		neVersion := ""

		if ne.PrimaryCurStatus == "ACTIVATED" {
			neVersion = ne.PrimarySWVersion
		} else if ne.SecondaryCurStatus == "ACTIVATED" {
			neVersion = ne.SecondarySWVersion
		} else {
			neVersion = "UNKNOWN"
		}
		rowToAdd := []string{
			fmt.Sprintf("%v", c),
			fmt.Sprintf("%v", neName),
			fmt.Sprintf("%v", neVersion),
		}
		output = append(output, rowToAdd)
		c++
	}
	return output
}

func updateSw(agent RestAgent) {

	nelist := neList(getNEs(agent))

	var wg sync.WaitGroup
	totalWorkers := 5
	busyWorkers := 0

	for _, ne := range nelist {

		wg.Add(1)
		busyWorkers += 1

		go func(ne NEs) {
			log.Printf("Reading the software version from %v.\n", ne.NeLabel)
			_ = agent.Get(fmt.Sprintf("/swim/neSwStatus?neLabel=%v&neType=--", ne.NeLabel), nil)
			wg.Done()
			busyWorkers -= 1
		}(ne)

		if busyWorkers >= totalWorkers {
			log.Println("Waiting for an idle worker.")
			wg.Wait()
		}
	}
	wg.Wait()
	output := exportPrep(getNEs(agent))
	err := exportFile(output)
	if err == nil {
		log.Println("SUCCESS: NE Software report file has been exported!")
	} else {
		panic(err)
	}
}

func timeCalculator() int64 {
	return time.Now().Unix()
}

func exportFile(output [][]string) error {
	ts := fmt.Sprintf("%v", timeCalculator())
	csvFile, err := os.Create(fmt.Sprintf("output_%v.csv", ts))
	if err != nil {
		return err
	} else {
		csvwriter := csv.NewWriter(csvFile)
		for _, empRow := range output {
			_ = csvwriter.Write(empRow)
		}
		csvwriter.Flush()
	}
	return nil
}

func main() {
	ipaddr := "1.1.1.1"
	uname := "test"
	passw := "******"
	restAgent := Init(ipaddr, uname, passw)
	defer restAgent.Logout()

	print(string(restAgent.Get("8443/oms1350/data/npr/nes", map[string]string{})))

}
