package main

import (
	"fmt"
	"bytes"
	"encoding/json"
	"strconv"
  "errors"
  "strings"
)

func queryLatest(src Puppet) ([]Latest, error) {
	response, err := queryMuxrpc(src.instanceID, "latest")
	if err != nil {
		return nil, err
	}
	responses := splitResponses(response.String())
	seqnos := make([]Latest, 0, len(responses))
	for _, str := range responses {
		var parsed Latest
		json.Unmarshal(bytes.NewBufferString(str).Bytes(), &parsed)
		seqnos = append(seqnos, parsed)
	}
	return seqnos, nil
}

func DoConnect(src, dst Puppet) error {
	portSrc := 18888 + src.instanceID
	portDst := 18888 + dst.instanceID
	dstMultiAddr := multiserverAddr(dst.feedID, portDst)

	CLI := "/home/cblgh/code/go/src/ssb/cmd/sbotcli/sbotcli"
	cmd := fmt.Sprintf(`%s -addr 192.168.88.18:%d --key /home/cblgh/code/netsim-experiments/ssb-server/puppet_%d/secret call conn.connect %s`, CLI, portSrc, src.instanceID, dstMultiAddr)
	response, err := runline(cmd)
	if err != nil {
		return err
	}
	taplog(response.String())
	return nil
}

func DoDisconnect(src, dst Puppet) error {
	portSrc := 18888 + src.instanceID
	portDst := 18888 + dst.instanceID
	dstMultiAddr := multiserverAddr(dst.feedID, portDst)

	CLI := "/home/cblgh/code/go/src/ssb/cmd/sbotcli/sbotcli"
	cmd := fmt.Sprintf(`%s -addr 192.168.88.18:%d --key /home/cblgh/code/netsim-experiments/ssb-server/puppet_%d/secret call conn.stop`, CLI, portSrc, src.instanceID)
	response, err := runline(cmd)
	if err != nil {
		return err
	}
	taplog(response.String())

	cmd = fmt.Sprintf(`%s -addr 192.168.88.18:%d --key /home/cblgh/code/netsim-experiments/ssb-server/puppet_%d/secret call conn.disconnect %s`, CLI, portSrc, src.instanceID, dstMultiAddr)
	response, err = runline(cmd)
	if err != nil {
		return err
	}
	taplog(response.String())
	return nil
}

// really bad Rammstein pun, sorry (absolutely not sorry)
func DoHast(src, dst Puppet, seqno string) error {
	srcLatestSeqs, err := queryLatest(src)
	if err != nil {
		return err
	}
	dstLatestSeqs, err := queryLatest(dst)
	if err != nil {
		return err
	}

	dstViaSrc := getLatestByFeedID(srcLatestSeqs, dst.feedID)
	dstViaDst := getLatestByFeedID(dstLatestSeqs, dst.feedID)

	var assertedSeqno int
	if seqno == "latest" {
		assertedSeqno = dstViaDst.Sequence
	} else {
		assertedSeqno, err = strconv.Atoi(seqno)
		if err != nil {
			m := fmt.Sprintf("expected keyword 'latest' or a numberd\nwas %s", seqno)
			return TestError{err: errors.New("sequence number wasn't a number (or latest)"), message: m}
		}
	}

	if dstViaSrc.Sequence == assertedSeqno && dstViaSrc.ID == dstViaDst.ID {
		return nil
	} else {
		m := fmt.Sprintf("expected sequence: %s at seq %d\nwas sequence %s at seq %d", dstViaDst.ID, assertedSeqno, dstViaSrc.ID, dstViaSrc.Sequence)
		return TestError{err: errors.New("sequences didn't match"), message: m}
	}
	return nil
}

func DoWhoami(instance int) (string, error) {
	response, err := queryMuxrpc(instance, "whoami")
	if err != nil {
		return "", err
	}
	var parsed Whoami
	json.Unmarshal(response.Bytes(), &parsed)
	return parsed.ID, nil
}

func DoLog(instance, n int) (string, error) {
	response, err := queryMuxrpc(instance, "log")
	if err != nil {
		return "", err
	}

	responses := splitResponses(response.String())
  length := len(responses)
  return strings.Join(responses[length-n:length], "\n"), nil
}

func DoFollow(instance int, feedID string, isFollow bool) error {
	var followType string
	if !isFollow { // => unfollow message
		followType = "no-"
	}
	followMsg := fmt.Sprintf(`publish --type contact --contact %s --%sfollowing`, feedID, followType)
	_, err := queryMuxrpc(instance, followMsg)
	if err != nil {
		return err
	}
	return nil
}

func queryIsFollowing(instance int, srcID, dstID string) (bool, error) {
	msg := fmt.Sprintf(`friends.isFollowing --source %s --dest %s`, srcID, dstID)
	res, err := queryMuxrpc(instance, msg)
	if err != nil {
		return false, err
	}
	isFollowing := strings.TrimSpace(res.String()) == "true"
  return isFollowing, nil
}

func DoIsFollowing(instance int, srcID, dstID string) error {
  isFollowing, err := queryIsFollowing(instance, srcID, dstID)
	if err != nil {
		return  err
	}
	if !isFollowing {
		m := fmt.Sprintf("%s did not follow %s", srcID, dstID)
		return TestError{err: errors.New("isfollowing returned false"), message: m}
	}
	return nil
}

func DoIsNotFollowing(instance int, srcID, dstID string) error {
  isFollowing, err := queryIsFollowing(instance, srcID, dstID)
	if err != nil {
		return  err
	}
	if isFollowing {
    m := fmt.Sprintf("%s should not follow %s\nactual: %s is following %s", srcID, dstID, srcID, dstID)
		return TestError{err: errors.New("isfollowing returned true"), message: m}
	}
	return nil
}

func DoPost(instance int) error {
	port := 18888 + instance
	CLI := "/home/cblgh/code/go/src/ssb/cmd/sbotcli/sbotcli"
	cmd := fmt.Sprintf(`%s -addr 192.168.88.18:%d --key /home/cblgh/code/netsim-experiments/ssb-server/puppet_%d/secret publish post bep`, CLI, port, instance)
	_, err := runline(cmd)
	if err != nil {
		return err
	}
	return nil
}

func DoLatest(instance int) error {
	postMsg := "latest"
	_, err := queryMuxrpc(instance, postMsg)
	if err != nil {
		return err
	}
	return nil
}
