package main

import (
	"encoding/json"
	"fmt"
	"time"
)

func pprint(content []map[string]interface{}) {
	for ind, item := range content {
		fmt.Printf("%v:\n", ind)
		for k, v := range item {
			fmt.Printf("%v : %v\n", k, v)
		}
	}
}

func updateSw(agent RestAgent) {

	var allNEsJson map[string][]map[string]interface{}
	allNEsRaw := agent.HttpGet("/data/swim/neSoftware", map[string]string{"Range": "items=0-1999"})
	json.Unmarshal([]byte(allNEsRaw), &allNEsJson)
	var neList []map[string]interface{}

	for _, ne := range allNEsJson["items"] {
		if ne["neType"] == "1830PSS-PHN" {
			neList = append(neList, ne)
		}
	}
	for _, ne := range neList {
		fmt.Printf("Reading the software version from %v.\n", ne["neLabel"])
		_ = agent.HttpGet(fmt.Sprintf("/swim/neSwStatus?neLabel=%v&neType=--", ne["neLabel"]), map[string]string{})
		time.Sleep(1 * time.Second)
	}
}

func GetRamanConnections(agent RestAgent) {

	rawOtsData := agent.HttpGet("/data/npr/physicalConns", map[string]string{})
	_, listJson := GeneralJsonDecoder(rawOtsData)
	var otslist []map[string]interface{}
	for _, phycon := range listJson {
		if phycon["wdmConnectionType"] == "WdmPortType_ots" {
			otslist = append(otslist, phycon)
		}
	}
	pprint(otslist)
}

func main() {
	ipaddr := "172.29.4.72"
	uname := "admin"
	passw := "Changeme_1@#"
	restAgent := Init(ipaddr, uname, passw)
	defer restAgent.NfmtDeauth()

	updateSw(restAgent)

}
