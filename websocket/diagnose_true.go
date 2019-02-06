// +build disgord_diagnosews

package websocket

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"time"
)

const SaveIncomingPackets = true

const DiagnosePath = "diagnose-report"
const DiagnosePath_packets = "diagnose-report/packets"

var outgoingPacketSequence uint = 0 // TODO: this needs to support sharding

func formatFilename(incoming bool, clientType int, shardID, opCode, sequencenr uint, suffix string) (filename string) {

	unix := strconv.FormatInt(time.Now().UnixNano(), 10)
	shard := strconv.FormatUint(uint64(shardID), 10)
	op := strconv.FormatUint(uint64(opCode), 10)
	seq := strconv.FormatUint(uint64(sequencenr), 10)

	var direction string
	if incoming {
		direction = "IN"
	} else {
		direction = "OUT"
	}

	t := "E"
	if clientType == clientTypeVoice {
		t = "V"
	}

	return unix + "_" + t + "_" + direction + "_id" + shard + "_op" + op + "_s" + seq + suffix + ".json"
}

// saveOutgoingPacket saves raw json content to disk
// format: I_<seq>_<op>_<shard_id>_<unix>.json
// unix is the unix timestamp on save
// seq is the sequence number: outgoingPacketSequence
// op is the operation code
func saveOutgoingPacket(c *client, packet *clientPacket) {
	data, err := json.MarshalIndent(packet, "", "\t")
	if err != nil {
		c.Debug(err)
	}

	filename := formatFilename(false, c.clientType, c.ShardID, packet.Op, outgoingPacketSequence, "")
	outgoingPacketSequence++

	path := DiagnosePath_packets + "/" + filename
	if err = ioutil.WriteFile(path, data, 0644); err != nil {
		c.Debug(err)
	}
}

// saveIncomingPacker saves raw json content to disk
// format: O_<unix>_<seq>_<op>_<shard_id>[_<evt_name>].json
// unix is the unix timestamp on save. This is needed as the sequence number can be reset.
// seq is the sequence number
// op is the operation code
// evt_name is the event name (optional)
func saveIncomingPacker(c *client, evt *DiscordPacket, packet []byte) {
	evtStr := "_" + evt.EventName
	if evtStr == "_" {
		evtStr = "_EMPTY"
	}

	filename := formatFilename(true, c.clientType, c.ShardID, evt.Op, evt.SequenceNumber, evtStr)
	path := DiagnosePath_packets + "/" + filename

	// pretty
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, packet, "", "\t"); err != nil {
		c.Debug(err)
	}

	if err := ioutil.WriteFile(path, prettyJSON.Bytes(), 0644); err != nil {
		c.Debug(err)
	}
}
