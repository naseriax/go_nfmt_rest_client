package main

import (
	"encoding/json"
	"fmt"
	"go_nfmt_rest_client/restlib"
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

func updateSw(agent restlib.RestAgent) {

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

func GetRamanConnections(agent restlib.RestAgent) {

	rawOtsData := agent.HttpGet("/data/npr/physicalConns", map[string]string{})
	_, listJson := restlib.GeneralJsonDecoder(rawOtsData)
	var otslist []map[string]interface{}
	for _, phycon := range listJson {
		if phycon["wdmConnectionType"] == "WdmPortType_ots" {
			otslist = append(otslist, phycon)
		}
	}
	pprint(otslist)
}

func main() {
	ipaddr := "1.1.1.1"
	uname := "admin"
	passw := "ppp"
	restAgent := restlib.Init(ipaddr, uname, passw)
	defer restAgent.NfmtDeauth()

	updateSw(restAgent)

}
