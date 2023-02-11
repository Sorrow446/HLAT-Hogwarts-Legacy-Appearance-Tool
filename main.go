package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
	_ "github.com/mattn/go-sqlite3"
)

var dbPath = filepath.Join(os.TempDir(), "hlse_tmp.db")

var magic = [4]byte{'\x47', '\x56', '\x41', '\x53'}

var rawDbImageStr = []byte{
	'\x52', '\x61', '\x77', '\x44', '\x61', '\x74',
	'\x61', '\x62', '\x61', '\x73', '\x65', '\x49',
	'\x6D', '\x61', '\x67', '\x65',
}

func extractDb(saveData []byte) (int, int, error) {
	imageStrStart := bytes.Index(saveData, rawDbImageStr)
	if imageStrStart == -1 {
		return 0, 0, errors.New("couldn't find db image string")
	}
	dbSizeOffset := imageStrStart+61
	dbStartOffset := dbSizeOffset+4
	dbSizeBytes := saveData[dbSizeOffset:dbStartOffset]
	dbSize := binary.LittleEndian.Uint32(dbSizeBytes)
	dbEndOffset := dbStartOffset+int(dbSize)
	dbData := saveData[dbStartOffset:dbEndOffset]
	err := os.WriteFile(dbPath, dbData, 0755)
	return imageStrStart, dbEndOffset, err
}

var resolveCommand = map[string]string{
	"export": "export",
	"e":      "export",
	"import": "import",
	"i":      "import",
}

var queries = map[string]string{
	"appearance_data":     `SELECT "PresetType", "PresetName" FROM "AvatarFullBodyPresetsDynamic" WHERE "RegistryId" = "Player0"`,
	"appearance_data_del": `DELETE FROM "AvatarFullBodyPresetsDynamic" WHERE "RegistryId" = "Player0"`,
	"appearance_data_ins": `INSERT INTO "AvatarFullBodyPresetsDynamic" ('RegistryId', 'PresetType', 'PresetName') VALUES ('Player0', '%s', '%s')`,
	"gender_data":         `SELECT "DataName", "DataValue" FROM "MiscDataDynamic" WHERE "DataOwner" = "Player" AND "DataName" IN("GenderPronoun", "GenderVoice", "GenderRig")`,
	"gender_data_updt":    `UPDATE "MiscDataDynamic" SET "DataValue" = "%s" WHERE "DataOwner" = "Player" AND "DataName" = "%s"`,
	"names":          	   `SELECT "DataValue" FROM "MiscDataDynamic" WHERE "DataName" = "%s"`,
	"names_updt":          `UPDATE "MiscDataDynamic" SET "DataValue" = "%s" WHERE "DataName" = "%s"`,
}


func parseArgs() (*Args, error) {
	var args Args
	arg.MustParse(&args)

	cmd, ok := resolveCommand[strings.ToLower(args.Command)]
	if !ok{
		return nil, errors.New("invalid command, must be import (i) or export (e)")
	}
	if cmd == "export" {
		if !(strings.HasSuffix(args.InPath, ".sav") && strings.HasSuffix(args.OutPath, ".json")) {
			return nil, errors.New("invalid input/output file extension")
		}
	} else {
		if !(strings.HasSuffix(args.InPath, ".json") && strings.HasSuffix(args.OutPath, ".sav")) {
			return nil, errors.New("invalid input/output file extension")
		}
	}
	args.Command = cmd
	return &args, nil
}

func updateRow(db *sql.DB, q string) error {
	res, err := db.Exec(q)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
	 	return err
	}
	if rowsAffected == 0 {
		return errors.New("db row wasn't updated")
	}
	return nil
}


func writeSave(updatedDbBytes, saveData []byte, imageStrStart, dbEndOffset int, outPath string) error {
	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(saveData[:imageStrStart+35])
	if err != nil {
		return err
	}
	buf := make([]byte, 4)


	updatedDbSize := len(updatedDbBytes)
	binary.LittleEndian.PutUint32(buf, uint32(updatedDbSize+4))

	_, err = f.Write(buf)
	if err != nil {
		return err
	}

	_, err = f.Write(saveData[imageStrStart+39:imageStrStart+61])
	if err != nil {
		return err
	}

	binary.LittleEndian.PutUint32(buf, uint32(updatedDbSize))

	_, err = f.Write(buf)
	if err != nil {
		return err
	}

	_, err = f.Write(updatedDbBytes)
	if err != nil {
		return err
	}
	_, err = f.Write(saveData[dbEndOffset:])
	return err
}

func parseAppearance(db *sql.DB) (*Appearance, error) {
	var (
		appearance Appearance
		presetType string
		presetName string
		name       string
	)

	for idx, dataName := range [2]string{"PlayerFirstName", "PlayerLastName"} {
		row := db.QueryRow(fmt.Sprintf(queries["names"], dataName))
		err := row.Err()
		if err != nil {
			return nil, err
		}
		err = row.Scan(&name)
		if err != nil {
			return nil, err
		}
		if idx == 0 {
			appearance.FirstName = name
		} else {
			appearance.LastName = name
		}
	}

	for idx, key := range [2]string{"appearance_data", "gender_data"} {
		rows, err := db.Query(queries[key])
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			err = rows.Scan(&presetType, &presetName)
			if err != nil {
				rows.Close()
				return nil, err
			}
			if idx == 0 {
				data := &AppearanceData{
					PresetType: presetType,
					PresetName: presetName,
				}
				appearance.AppearanceData = append(appearance.AppearanceData, data)
			} else {
				data := &GenderData{
					DataValue: presetName,
					DataName: presetType,
				}
				appearance.GenderData = append(appearance.GenderData, data)
			}
		}
		rows.Close()
	}
	return &appearance, nil

}

func readJsonApp(appPath string) (*Appearance, error) {
	data, err := os.ReadFile(appPath)
	if err != nil {
		return nil, err
	}
	var obj Appearance
	err = json.Unmarshal(data, &obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

func writeJsonApp(outPath string, appearance *Appearance) error {
	m, err := json.MarshalIndent(appearance, "", "\t")
 	if err != nil {
 		return err
 	}
 	fmt.Println(string(m))
	err = os.WriteFile(outPath, m, 0755)
	return err
}

func importApp(db *sql.DB, saveData []byte, imageStrStart, dbEndOffset int, args *Args) error {
	appearance, err := readJsonApp(args.InPath)
	if err != nil {
		db.Close()
		return err
	}

	if !args.OrigName {
		err = updateRow(db, fmt.Sprintf(queries["names_updt"], appearance.FirstName, "PlayerFirstName"))
		if err != nil {
			db.Close()
			return err
		}

		err = updateRow(db, fmt.Sprintf(queries["names_updt"], appearance.LastName, "PlayerLastName"))
		if err != nil {
			db.Close()
			return err
		}
	}

	err = updateRow(db, queries["appearance_data_del"])
	if err != nil {
		db.Close()
		return err
	}

	for _, d := range appearance.AppearanceData {
		err = updateRow(db, fmt.Sprintf(queries["appearance_data_ins"], d.PresetType, d.PresetName))
		if err != nil {
			db.Close()
			return err
		}
	}

	for _, d := range appearance.GenderData {
		err = updateRow(db, fmt.Sprintf(queries["gender_data_updt"], d.DataValue, d.DataName))
		if err != nil {
			db.Close()
			return err
		}
	}
	db.Close()
	updatedDbBytes, err := os.ReadFile(dbPath)
	if err != nil {
		return err
	}

	err = writeSave(updatedDbBytes, saveData, imageStrStart, dbEndOffset, args.OutPath)
	return err
}

func exportApp(db *sql.DB, outPath string) error {
	defer db.Close()
	appearance, err := parseAppearance(db)
	if err != nil {
		return err
	}
	err = writeJsonApp(outPath, appearance)
	return err
}

func main() {
	args, err := parseArgs()
	if err != nil {
		panic(err)
	}

	var saveData []byte

	if args.Command == "export" {
		saveData, err = os.ReadFile(args.InPath)
	} else {
		saveData, err = os.ReadFile(args.OutPath)
	}
	
	if err != nil {
		panic(err)
	}

	if !bytes.Equal(saveData[:4], magic[:]) {
		panic("invalid save file magic")
	}

	imageStrStart, dbEndOffset, err := extractDb(saveData)
	if err != nil {
		panic(err)
	}

	defer os.Remove(dbPath)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	switch args.Command {
	case "export":
		err = exportApp(db, args.OutPath)
	case "import":
		err = importApp(db, saveData, imageStrStart, dbEndOffset, args)
	}
	if err != nil {
		panic(err)
	}
}