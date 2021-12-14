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

func getNEs(agent RestAgent) map[string][]map[string]interface{} {
	var allNEsJson map[string][]map[string]interface{}
	allNEsRaw := agent.HttpGet("/data/swim/neSoftware", map[string]string{"Range": "items=0-1999"})
	json.Unmarshal([]byte(allNEsRaw), &allNEsJson)

	return allNEsJson
}

func neList(allNEsJson map[string][]map[string]interface{}) []map[string]interface{} {
	var neList []map[string]interface{}
	for _, ne := range allNEsJson["items"] {
		if ne["neType"] == "1830PSS-PHN" {
			neList = append(neList, ne)
		}
	}
	return neList
}

func updateSw(agent RestAgent) {

	allNEsJson := getNEs(agent)

	nelist := neList(allNEsJson)

	var wg sync.WaitGroup
	workers := 5
	j := 0
	for _, ne := range nelist {
		if j >= workers {
			fmt.Println("Waiting for an idle worker.")
			wg.Wait()
		}

		wg.Add(1)
		j += 1

		go func(ne map[string]interface{}) {
			fmt.Printf("Reading the software version from %v.\n", ne["neLabel"])
			_ = agent.HttpGet(fmt.Sprintf("/swim/neSwStatus?neLabel=%v&neType=--", ne["neLabel"]), nil)
			wg.Done()
			j -= 1
		}(ne)
	}
	wg.Wait()

	output := [][]string{
		{
			"No",
			"NE Name",
			"Software Version",
		},
	}
	c := 1
	allNEsJson = getNEs(agent)
	neVersion := ""

	for _, ne := range allNEsJson["items"] {
		neName := ne["neLabel"]

		if ne["primaryCurStatus"] == "ACTIVATED" {
			neVersion = ne["primarySWVersion"].(string)
		} else if ne["secondaryCurStatus"] == "ACTIVATED" {
			neVersion = ne["secondarySWVersion"].(string)
		} else {
			neVersion = "UNKNOWN"
		}

		rowToAdd := []string{
			fmt.Sprintf("%v", c),
			fmt.Sprintf("%v", neName),
			fmt.Sprintf("%v", neVersion),
		}
		output = append(output, rowToAdd)
	}

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
	ipaddr := "1.2.3.4"
	uname := "alcatel"
	passw := "password"
	restAgent := Init(ipaddr, uname, passw)
	defer restAgent.NfmtDeauth()

	updateSw(restAgent)

}
