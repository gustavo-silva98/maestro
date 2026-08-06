package orchestrator

import (
	"fmt"
	"os"
	"testing"
)

func TestBuildVolumes(t *testing.T) {
	t.Run("Build String sem erros", func(t *testing.T) {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatalf("erro ao descobrir diretório")
		}
		expOutputDir := fmt.Sprintf("%s/scripts/football/output-teste:/app/Prints-teste", cwd)
		expInputFile := fmt.Sprintf("%s/scripts/football/input-teste.csv:/app/input-teste.csv", cwd)
		inputFile, outputDir, err := buildVolumes("teste")
		if err != nil {
			t.Fatalf("Erro ao gerar string de volumes: %v", err)
		}
		if inputFile != expInputFile {
			t.Errorf("erro ao gerar string de inputFile - Exp:%v - Rec: %v", expInputFile, inputFile)
		}
		if outputDir != expOutputDir {
			t.Errorf("erro ao gerar string de outputDir - Exp:%v - Rec: %v", expOutputDir, outputDir)
		}
	})
}
