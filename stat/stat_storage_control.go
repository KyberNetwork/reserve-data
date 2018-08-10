package stat

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
)

func (self ReserveStats) ControllPriceAnalyticSize() error {
	for {
		log.Printf("StatPruner: waiting for signal from analytic storage control channel")
		t := <-self.storageController.Runner.GetAnalyticStorageControlTicker()
		timepoint := common.TimeToTimepoint(t)
		log.Printf("StatPruner: got signal in analytic storage control channel with timestamp %d", timepoint)
		fileName := fmt.Sprintf("./exported/ExpiredPriceAnalyticData_%s", time.Unix(int64(timepoint/1000), 0).UTC())
		nRecord, err := self.analyticStorage.ExportExpiredPriceAnalyticData(common.GetTimepoint(), fileName)
		if err != nil {
			log.Printf("ERROR: StatPruner export Price Analytic operation failed: %s", err)
		} else {
			var integrity bool
			if nRecord > 0 {
				err = self.storageController.Arch.UploadFile(self.storageController.Arch.GetStatDataBucketName(), self.storageController.ExpiredPriceAnalyticPath, fileName)
				if err != nil {
					log.Printf("StatPruner: Upload file failed: %s", err)
				} else {
					integrity, err = self.storageController.Arch.CheckFileIntergrity(self.storageController.Arch.GetStatDataBucketName(), self.storageController.ExpiredPriceAnalyticPath, fileName)
					if err != nil {
						log.Printf("ERROR: StatPruner: error in file integrity check (%s):", err)
					}
					if !integrity {
						log.Printf("ERROR: StatPruner: file upload corrupted")

					}
					if err != nil || !integrity {
						//if the intergrity check failed, remove the remote file.
						removalErr := self.storageController.Arch.RemoveFile(self.storageController.Arch.GetStatDataBucketName(), self.storageController.ExpiredPriceAnalyticPath, fileName)
						if removalErr != nil {
							log.Printf("ERROR: StatPruner: cannot remove remote file :(%s)", removalErr)
						}
					}
				}
			}
			if integrity && err == nil {
				nPrunedRecords, err := self.analyticStorage.PruneExpiredPriceAnalyticData(common.TimeToTimepoint(t))
				if err != nil {
					log.Printf("StatPruner: cannot prune Price Analytic Data (%s)", err)
				} else if nPrunedRecords != nRecord {
					log.Printf("StatPruner: Number of exported Data is %d, which is different from number of Pruned Data %d", nRecord, nPrunedRecords)
				} else {
					log.Printf("StatPruner: exported and pruned %d expired records from Price Analytic Data", nRecord)
				}
			}
		}
		if err := os.Remove(fileName); err != nil {
			log.Fatal(err)
		}
	}
}

// uploadAndVerify upload the file to remote storage and check its integrity
// return error if occur and backup integrity

func (self ReserveStats) uploadAndVerify(fileName, remotePath string) (bool, error) {
	err := self.storageController.Arch.UploadFile(self.storageController.Arch.GetStatDataBucketName(), remotePath, fileName)
	if err != nil {
		return false, err
	}

	integrity, err := self.storageController.Arch.CheckFileIntergrity(self.storageController.Arch.GetStatDataBucketName(), remotePath, fileName)
	if err != nil {
		log.Printf("ERROR: StatPruner: error in file integrity check (%s):", err)
		return false, err
	}

	//if the integrity check doesn't meet any error but the integrity is false, remove it from remote storage
	if !integrity {
		log.Printf("ERROR: StatPruner: file upload corrupted")
		removalErr := self.storageController.Arch.RemoveFile(self.storageController.Arch.GetStatDataBucketName(), remotePath, fileName)
		if removalErr != nil {
			log.Printf("ERROR: StatPruner: cannot remove remote file :(%s)", removalErr)
		}
		return integrity, removalErr
	}
	return true, nil
}

// ControlRateSize will check the rate database and export all the record that is more than 30 days old from now
// It will export these record to mutiple gz file, each contain all the expired record in a certain date
// in format ExpiredRateData_<firstTimestamp>_<lastTimestmap>.gz, in which time stamp is second
func (self ReserveStats) ControlRateSize() error {
	for {
		t := <-self.storageController.Runner.GetRateStorageControlTicker()
		timepoint := common.TimeToTimepoint(t)
		log.Printf("StatPruner: got signal in rate storage control channel with timestamp %d", timepoint)
		//continuously pruning until there is no more expired data.
		for {
			tempfileName := fmt.Sprintf("./exported/TempExpireRateData.gz")
			fromTime, toTime, nRecord, err := self.rateStorage.ExportExpiredRateData(common.GetTimepoint(), tempfileName)
			fileName := fmt.Sprintf("./exported/ExpiredRateData_%d_%d.gz", fromTime/1000, toTime/1000)
			if rErr := os.Rename(tempfileName, fileName); rErr != nil {
				log.Printf("StatPruner: cannot rename file (%s)", rErr)
				break
			}
			if err != nil {
				log.Printf("StatPruner ERROR: StatPruner export Price Analytic operation failed: %s", err)
				break
			} else {
				if nRecord > 0 {
					integrity, err := self.uploadAndVerify(fileName, self.storageController.ExpiredRatePath)
					if integrity && err == nil {
						nPrunedRecords, err := self.rateStorage.PruneExpiredReserveRateData(toTime)
						if err != nil {
							log.Printf("StatPruner: cannot prune Reserve rate Data (%s)", err)
						} else if nPrunedRecords != nRecord {
							log.Printf("StatPruner: Number of exported Data is %d, which is different from number of Pruned Data %d", nRecord, nPrunedRecords)
						} else {
							log.Printf("StatPruner: exported and pruned %d expired records from Reserve rate Data", nRecord)
						}
					}
				} else {
					log.Printf("StatPruner: No record, exit and wait for next timestamp")
					//if there is no expired record, break this loop
					break
				}

			}
			// remove the temp file
			if err := os.Remove(fileName); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (self ReserveStats) RunStorageController() error {
	err := self.storageController.Runner.Start()
	if err != nil {
		return err
	}
	go func() {
		if cErr := self.ControllPriceAnalyticSize(); cErr != nil {
			log.Printf("Control price analytic failed: %s", cErr.Error())
		}
	}()
	go func() {
		if rErr := self.ControlRateSize(); rErr != nil {
			log.Printf("StatPruner: Control rate analytic failed: %s", rErr.Error())
		}
	}()
	return err
}
