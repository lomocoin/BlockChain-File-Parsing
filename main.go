package main

import (
	"log"
	"github.com/lomocoin/blockchain-parsing/lomocoin"
	"github.com/lomocoin/blockchain-parsing/lib/others/blockdb"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/lomocoin/blockchain-parsing/events"
	"encoding/csv"
	"strconv"
	"os"
	"github.com/therecipe/qt/widgets"
	"fmt"
	"os/user"
)

func main() {
	widgets.NewQApplication(len(os.Args), os.Args)

	var (
		echoGroup          = widgets.NewQGroupBox2("BlockFile Path", nil)
		blockChainFilePath = widgets.NewQLineEdit(nil)
	)
	blockChainFilePath.SetPlaceholderText("Enter the block file path")

	var (
		inputMaskGroup   = widgets.NewQGroupBox2("Block LevelDB Path", nil)
		blockLeveldbPath = widgets.NewQLineEdit(nil)
	)
	blockLeveldbPath.SetPlaceholderText("Enter the leveldb path")

	var (
		blockHeightIndexGroup = widgets.NewQGroupBox2("Block Height", nil)
		blockStartIndex       = widgets.NewQLineEdit(nil)
		blockEndIndex         = widgets.NewQLineEdit(nil)
	)
	blockStartIndex.SetPlaceholderText("Enter the start block height")
	blockEndIndex.SetPlaceholderText("Enter the end block height")

	var (
		eventState         = 2
		radioButtonGrop    = widgets.NewQButtonGroup(nil)
		radioButtonBox     = widgets.NewQGroupBox2("Event", nil)
		radioButtonCsv     = widgets.NewQRadioButton2("CSV", nil)
		radioButtonConsole = widgets.NewQRadioButton2("Console", nil)
	)
	radioButtonConsole.SetChecked(true)
	radioButtonGrop.AddButton(radioButtonCsv, 1)
	radioButtonGrop.AddButton(radioButtonConsole, 2)

	var textEditConsole = widgets.NewQTextEdit(nil)
	textEditConsole.SetPlaceholderText("Console log......")
	textEditConsole.SetMaximumHeight(300)

	var echoLayout = widgets.NewQGridLayout(nil)
	echoLayout.AddWidget3(blockChainFilePath, 1, 0, 1, 2, 0)
	echoGroup.SetLayout(echoLayout)
	echoGroup.SetMaximumHeight(100)

	var inputMaskLayout = widgets.NewQGridLayout(nil)
	inputMaskLayout.AddWidget3(blockLeveldbPath, 1, 0, 1, 2, 0)
	inputMaskGroup.SetLayout(inputMaskLayout)
	inputMaskGroup.SetMaximumHeight(100)

	var radioLayout = widgets.NewQGridLayout(nil)
	radioLayout.AddWidget(radioButtonCsv, 0, 0, 0)
	radioLayout.AddWidget(radioButtonConsole, 0, 1, 0)
	radioButtonBox.SetLayout(radioLayout)
	radioButtonBox.SetMaximumHeight(100)

	var blockHeightLayout = widgets.NewQGridLayout(nil)
	blockHeightLayout.AddWidget(blockStartIndex, 0, 0, 0)
	blockHeightLayout.AddWidget(blockEndIndex, 0, 1, 0)
	blockHeightIndexGroup.SetLayout(blockHeightLayout)
	blockHeightIndexGroup.SetMaximumHeight(100)

	radioButtonGrop.ConnectButtonClicked2(func(id int) {
		eventState = id;
	})

	var runButton = widgets.NewQPushButton2("RUN", nil)
	runButton.ConnectClicked(func(checked bool) {
		startApplication(blockChainFilePath.Text(), blockLeveldbPath.Text(), blockStartIndex.Text(), blockEndIndex.Text(), eventState, textEditConsole)
	})

	var layout = widgets.NewQGridLayout2()
	layout.AddWidget(echoGroup, 0, 0, 0)
	layout.AddWidget(inputMaskGroup, 0, 1, 0)
	layout.AddWidget3(radioButtonBox, 1, 0, 1, 2, 0)
	layout.AddWidget3(blockHeightIndexGroup, 2, 0, 1, 2, 0)
	layout.AddWidget3(runButton, 3, 0, 1, 2, 0)
	layout.AddWidget3(textEditConsole, 4, 0, 1, 2, 0)

	var window = widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle("LomoCoin BlockChain Parsing")
	window.SetMaximumWidth(800)
	window.SetMinimumWidth(800)
	window.SetMaximumHeight(650)
	window.SetMinimumHeight(650)

	var centralWidget = widgets.NewQWidget(window, 0)
	centralWidget.SetLayout(layout)
	window.SetCentralWidget(centralWidget)

	window.Show()

	widgets.QApplication_Exec()
}

func startApplication(blockFilePath string, levelPath string, startHeight string, endHeight string, state int, console *widgets.QTextEdit) {
	Magic := [4]byte{0xa6, 0xb8, 0xc9, 0xd5}

	try(func() {
		BlockDatabase := blockdb.NewBlockDB(blockFilePath, Magic)

		db, leveldbErr := leveldb.OpenFile(levelPath, nil)
		if leveldbErr != nil {
			panic("Leveldb failed to open, please check the path Or delete the LOCK file under the leveldb directory.")
		}

		start_block, _ := strconv.Atoi(startHeight)
		end_block, _ := strconv.Atoi(endHeight)

		if start_block < 1 {
			panic("starting height must be greater than 0")
		}

		var writeObj *csv.Writer

		// csv
		if state == 1 {
			user, userError := user.Current()

			if nil == userError {

				file, err := os.Create(user.HomeDir + "/Desktop/block.csv")
				if err != nil {
					panic(err)
				}
				defer file.Close()

				file.WriteString("\xEF\xBB\xBF")

				writeObj = csv.NewWriter(file)
				writeObj.Write([]string{"block_index", "block_hash", "tx_time", "tx_id", "tx_in", "tx_out"})
			} else {
				panic(userError)
			}
		}

		for blockIndex := 0; blockIndex <= end_block; blockIndex++ {
			dat, er := BlockDatabase.FetchNextBlock()
			if dat == nil || er != nil {
				log.Println("END of DB file")
				break
			}
			bl, er := lomocoin.NewBlock(dat[:])

			if er != nil {
				println("Block inconsistent:", er.Error())
				break
			}

			if blockIndex < start_block {
				continue
			}

			dbBlockHeight := bl.LeveldbFindBlockHeightWhereHash(db, bl.Hash.Hash)

			// Fork block
			if blockIndex != dbBlockHeight {
				blockIndex = dbBlockHeight
			}

			console.InsertPlainText(fmt.Sprintf("Current block height: %v \n", blockIndex))

			bl.BuildTxList()

			if state == 1 {
				events.WriteCSV(bl, blockIndex, writeObj)
			}

			if state == 2 {
				events.Put(bl, console)
			}
		}

		os.Remove(levelPath+"/LOCK")

		if state == 1 {
			console.InsertPlainText("\nFile export completion!\n")
		}

	}, func(er interface{}) {
		console.InsertPlainText(fmt.Sprintf("\nError: %v \n", er))
	})
}

func try(function func(), handler func(interface{})) {
	defer func() {
		if error := recover(); error != nil {
			handler(error)
		}
	}()

	function()
}
