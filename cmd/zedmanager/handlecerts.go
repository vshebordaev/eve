// Copyright (c) 2017 Zededa, Inc.
// All rights reserved.

package main

import (
	"github.com/zededa/go-provision/types"
	"log"
)

// Key is UUID
var certObjStatus map[string]types.CertObjStatus

func handleCertObjStatusModify(ctxArg interface{}, statusFilename string,
	statusArg interface{}) {
	status := statusArg.(*types.CertObjStatus)
	if status == nil {
		return
	}
	uuidStr := status.UUIDandVersion.UUID.String()

	log.Printf("handlCertObjStatusModify for %s\n", uuidStr)

	if certObjStatus == nil {
		log.Printf("create CertObj Status map\n")
		certObjStatus = make(map[string]types.CertObjStatus)
	}

	changed := false
	if m, ok := certObjStatus[uuidStr]; ok {
		if status.State != m.State {
			if debug {
				log.Printf("Cert obj status map changed from %v to %v\n",
					m.State, status.State)
			}
			changed = true
		}
	} else {
		if debug {
			log.Printf("Cert objmap add for %v\n", status.State)
		}
		changed = true
	}
	if changed {
		certObjStatus[uuidStr] = *status
		updateAIStatusUUID(uuidStr)
	}

	log.Printf("handleCertObjrStatusModify done for %s\n", uuidStr)
}

func handleCertObjStatusDelete(ctxArg interface{}, statusFilename string) {
	log.Printf("handleCertObjtatusDelete for %s\n", statusFilename)

	key := statusFilename
	if m, ok := certObjStatus[key]; !ok {
		log.Printf("handleCertObjStatusDelete for %s - not found\n",
			key)
	} else {
		state := m.State
		delete(certObjStatus, key)
		log.Printf("handleCertObjStatusDelete done for %s, in %v\n",
			statusFilename, state)
	}
}
