// Copyright (c) 2017 Zededa, Inc.
// All rights reserved.

// zboot APIs for IMGA  & IMGB

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zededa/go-provision/types"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	tmpDir        = "/var/tmp/zededa"
	imgAPartition = tmpDir + "/IMGAPart"
	imgBPartition = tmpDir + "/IMGBPart"
)

// reset routine
func zbootReset() {
	rebootCmd := exec.Command("zboot", "reset")
	_, err := rebootCmd.Output()
	if err != nil {
		log.Fatalf("zboot reset: err %v\n", err)
	}
}

// partition routines
func getCurrentPartition() string {
	curPartCmd := exec.Command("zboot", "curpart")
	ret, err := curPartCmd.Output()
	if err != nil {
		log.Fatalf("zboot curpart: err %v\n", err)
	}

	partName := string(ret)
	partName = strings.TrimSpace(partName)
	return partName
}

func getOtherPartition() string {

	partName := getCurrentPartition()

	switch partName {
	case "IMGA":
		partName = "IMGB"
	case "IMGB":
		partName = "IMGA"
	default:
		log.Fatalf("getOtherPartition unknow partName %s\n", partName)
	}
	return partName
}

func validatePartitionName(partName string) {

	if partName == "IMGA" || partName == "IMGB" {
		return
	}
	errStr := fmt.Sprintf("invalid partition %s", partName)
	log.Fatal(errStr)
}

func validatePartitionState(partState string) {
	if partState == "active" || partState == "inprogress" ||
		partState == "unused" || partState == "updating" {
		return
	}
	errStr := fmt.Sprintf("invalid partition state %s", partState)
	log.Fatal(errStr)
}

func isCurrentPartition(partName string) bool {
	validatePartitionName(partName)
	curPartName := getCurrentPartition()
	return curPartName == partName
}

func isOtherPartition(partName string) bool {
	validatePartitionName(partName)
	otherPartName := getOtherPartition()
	return otherPartName == partName
}

//  get/set api routines
func getPartitionState(partName string) string {

	validatePartitionName(partName)

	partStateCmd := exec.Command("zboot", "partstate", partName)
	ret, err := partStateCmd.Output()
	if err != nil {
		log.Fatalf("zboot partstate %s: err %v\n", partName, err)
	}
	partState := string(ret)
	partState = strings.TrimSpace(partState)
	log.Printf("zboot partstate %s: %v\n", partName, partState)
	return partState
}

func isPartitionState(partName string, partState string) bool {

	validatePartitionName(partName)
	validatePartitionState(partState)

	curPartState := getPartitionState(partName)
	log.Printf("isPartitionState %s: %v %v\n",
		partName, curPartState, partState)
	return curPartState == partState
}

func setPartitionState(partName string, partState string) {

	validatePartitionName(partName)
	validatePartitionState(partState)

	setPartStateCmd := exec.Command("zboot", "set_partstate",
		partName, partState)
	if _, err := setPartStateCmd.Output(); err != nil {
		log.Fatalf("zboot set_partstate %s %s: err %v\n",
			partName, partState, err)
	}
}

func getPartitionDevname(partName string) string {

	validatePartitionName(partName)
	getPartDevCmd := exec.Command("zboot", "partdev", partName)
	ret, err := getPartDevCmd.Output()
	if err != nil {
		log.Fatalf("zboot partdev %s: err %v\n", partName, err)
	}

	devName := string(ret)
	devName = strings.TrimSpace(devName)
	return devName
}

// set routines
func setPartitionStateActive(partName string) {
	setPartitionState(partName, "active")
}

func setPartitionStateUnused(partName string) {
	setPartitionState(partName, "unused")
}

func setPartitionStateUpdating(partName string) {
	setPartitionState(partName, "updating")
}

// check routines, for current partition
func isCurrentPartitionStateActive() bool {
	partName := getCurrentPartition()
	return isPartitionState(partName, "active")
}

func isCurrentPartitionStateInProgress() bool {
	partName := getCurrentPartition()
	return isPartitionState(partName, "inprogress")
}

func isCurrentPartitionStateUpdating() bool {
	partName := getCurrentPartition()
	return isPartitionState(partName, "updating")
}

// check routines, for other partition
func isOtherPartitionStateActive() bool {
	partName := getOtherPartition()
	return isPartitionState(partName, "active")
}

func isOtherPartitionStateInProgress() bool {
	partName := getOtherPartition()
	return isPartitionState(partName, "inprogress")
}

func isOtherPartitionStateUnused() bool {
	partName := getOtherPartition()
	return isPartitionState(partName, "unused")
}

func isOtherPartitionStateUpdating() bool {
	partName := getOtherPartition()
	return isPartitionState(partName, "updating")
}

func setCurrentPartitionStateActive() {
	partName := getCurrentPartition()
	setPartitionState(partName, "active")
}

func setCurrentPartitionStateUpdating() {
	partName := getCurrentPartition()
	setPartitionState(partName, "updating")
}

func setCurrentPartitionStateUnused() {
	partName := getCurrentPartition()
	setPartitionState(partName, "unused")
}

// set routines, for other partition
func setOtherPartitionStateActive() {
	partName := getOtherPartition()
	setPartitionState(partName, "active")
}

func setOtherPartitionStateUpdating() {
	partName := getOtherPartition()
	setPartitionState(partName, "updating")
}

func setOtherPartitionStateUnused() {
	partName := getOtherPartition()
	setPartitionState(partName, "unused")
}

func getCurrentPartitionDevName() string {
	partName := getCurrentPartition()
	return getPartitionDevname(partName)
}

func getOtherPartitionDevName() string {
	partName := getOtherPartition()
	return getPartitionDevname(partName)
}

func zbootWriteToPartition(srcFilename string, partName string) error {

	log.Printf("WriteToPartition %s: %s\n", partName, srcFilename)

	if ret := isOtherPartition(partName); ret == false {
		errStr := fmt.Sprintf("%s: not other partition", partName)
		log.Println(errStr)
		return errors.New(errStr)
	}

	if !isOtherPartitionStateUnused() {
		errStr := fmt.Sprintf("%s: Not an unused partition", partName)
		log.Println(errStr)
		return errors.New(errStr)
	}

	devName := getPartitionDevname(partName)
	if devName == "" {
		errStr := fmt.Sprintf("null devname for partition %s", partName)
		log.Println(errStr)
		return errors.New(errStr)
	}

	ddCmd := exec.Command("dd", "if="+srcFilename, "of="+devName, "bs=8M")
	if _, err := ddCmd.Output(); err != nil {
		log.Printf("partName : %v\n", err)
		return err
	}
	return nil
}

func partitionInit() {

	curPart := getCurrentPartition()
	otherPart := getOtherPartition()

	currActiveState := isCurrentPartitionStateActive()
	otherActiveState := isOtherPartitionStateActive()

	if currActiveState && otherActiveState {
		log.Printf("Both partitions are Active %s, %s n", curPart, otherPart)
		log.Printf("Mark other partition %s, unused\n", otherPart)
		setOtherPartitionStateUnused()
	}
}

func markPartitionStateActive() error {

	curPart := getCurrentPartition()
	otherPart := getOtherPartition()

	log.Printf("Check current partition %s, for inProgress state\n", curPart)
	if ret := isCurrentPartitionStateInProgress(); ret == false {
		errStr := fmt.Sprintf("Current partition %s, is not inProgress",
			curPart)
		return errors.New(errStr)
	}

	log.Printf("Mark the current partition %s, active\n", curPart)
	setCurrentPartitionStateActive()

	log.Printf("Check other partition %s for active state\n", otherPart)
	if ret := isOtherPartitionStateActive(); ret == false {
		errStr := fmt.Sprintf("Other partition %s, is not active",
			otherPart)
		return errors.New(errStr)
	}

	log.Printf("Mark other partition %s, unused\n", otherPart)
	setOtherPartitionStateUnused()
	return nil
}

// Partition Map Management routines
func readPartitionInfo(partName string) *types.PartitionInfo {

	validatePartitionName(partName)

	mapFilename := configDir + "/" + partName + ".json"
	if _, err := os.Stat(mapFilename); err != nil { 
		return nil
	}

	bytes, err := ioutil.ReadFile(mapFilename)
	if err != nil {
		return nil
	}

	partInfo := &types.PartitionInfo{}
	if err := json.Unmarshal(bytes, partInfo); err != nil {
		return nil
	}
	return partInfo
}

func readCurrentPartitionInfo() *types.PartitionInfo {
	partName := getCurrentPartition()
	return readPartitionInfo(partName)
}

func readOtherPartitionInfo() *types.PartitionInfo {
	partName := getOtherPartition()
	return readPartitionInfo(partName)
}

// haas to be always other partition
func removePartitionMap(mapFilename string, partInfo *types.PartitionInfo) error {
	otherPartInfo, err := readOtherPartitionInfo()
	if err != nil {
		return
	}

	// if same UUID, return
	if partInfo != nil &&
		partInfo.UUIDandVersion == otherPartInfo.UUIDandVersion {
		return nil
	}

	// old map entry, nuke it
	uuidStr := otherPartInfo.UUIDandVersion.UUID.String()

	// find the baseOs config/status map entries

	// reset the partition information
	config := baseOsConfigGet(uuidStr)
	if config != nil {
		configFilename := zedagentBaseOsConfigDirname + 
			"/" + uuidStr + ".json"
		config.PartitionLabel = ""
		for _, sc := range config.StorageConfigList {
			sc.FinalObjDir = ""
		}
		writeBaseOsConfig(*config, configFilename)
	}

	// and mark status as DELIVERED
	status := baseOsStatusGet(uuidStr)
	if status != nil {
		statusFilename := zedagentBaseOsStatusDirname + 
			"/" + uuidStr + ".json"
		status.State = types.DELIVERED
		errStr := fmt.Sprintf("uninstalled from %s",
			 otherPartInfo.PartitionLabel)
		status.Error = errStr
		status.ErrorTime = time.Now()
		writeBaseOsStatus(status, statusFilename)
	}

	partMapFilename := configDir + "/" + otherPartInfo.PartitionLabel + ".json"
	if err := os.Remove(partMapFilename); err != nil {
		log.Printf("%v for %s\n", err, partMapFilename)
		return
	}
	return
}

// check the partition table, for this baseOs
func getPersistentPartitionInfo(uuidStr string, imageSha256 string) string {

	var isCurrentPart, isOtherPart bool

	if partInfo := readCurrentPartitionInfo(); partInfo != nil {
		curUuidStr := partInfo.UUIDandVersion.UUID.String()
		if curUuidStr == uuidStr {
			isCurrentPart = true
		} else {
			if imageSha256 != "" &&
				imageSha256 == partInfo.ImageSha256 {
				isCurrentPart = true
			}
		}
	}

	if partInfo := readOtherPartitionInfo(); partInfo != nil {
		otherUuidStr := partInfo.UUIDandVersion.UUID.String()
		if otherUuidStr == uuidStr {
			isOtherPart = true
		} else {
			if imageSha256 != "" &&
				imageSha256 == partInfo.ImageSha256 {
				isOtherPart = true
			}
		}
	}

	if isCurrentPart == true && 
		isCurrentPart == isOtherPart {
		log.Fatal("Both partitions assigned with the same BaseOs %s\n", uuidStr)
	}

	if isCurrentPart == true {
		return getCurrentPartition()
	}

	if isOtherPart == true {
		return getOtherPartition()
	}
	return ""
}

// can only be done to the other partition
func setPersistentPartitionInfo(uuidStr string, config types.BaseOsConfig) error {
	partName := config.PartitionLabel
	log.Printf("%s, set partition %s\n", uuidStr, partName)

	if ret := isOtherPartition(partName); ret == false {
		errStr := fmt.Sprintf("%s: not other partition", partName)
		log.Println(errStr)
		return errors.New(errStr)
	}

	// new partition mapping
	partInfo := &types.PartitionInfo{}
	partInfo.UUIDandVersion = config.UUIDandVersion
	partInfo.ImageSha256    = getBaseOsImageSha(config)
	partInfo.BaseOsVersion  = config.BaseOsVersion
	partInfo.PartitionLabel = partName

	// remove old partition mapping
	mapFilename := configDir + "/" + config.PartitionLabel + ".json"
	removePartitionMap(mapFilename, partInfo)

	bytes, err := json.Marshal(partInfo)
	if  err != nil {
		errStr := fmt.Sprintf("%s, marshalling error %s\n", uuidStr, err)
		log.Println(errStr)
		return errors.New(errStr)
	}

	if err := ioutil.WriteFile(mapFilename, bytes, 0644); err != nil {
		errStr := fmt.Sprintf("%s, file write error %s\n", uuidStr, err)
		log.Println(errStr)
		return errors.New(errStr)
	}
	return nil
}

func resetPersistentPartitionInfo(uuidStr string) error {

	log.Printf("%s, reset partition\n", uuidStr)
	config := baseOsConfigGet(uuidStr)
	if config == nil {
		errStr := fmt.Sprintf("%s, config absent\n", uuidStr)
		err := errors.New(errStr)
		return err 
	}

	if !isOtherPartition(config.PartitionLabel) {
		return nil
	}
	mapFilename := configDir + "/" + config.PartitionLabel + ".json"
	return removePartitionMap(mapFilename, nil)
}
