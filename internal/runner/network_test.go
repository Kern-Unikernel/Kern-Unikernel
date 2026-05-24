package runner

import (
	"os"
	"testing"
)

func TestCreateTAPInterfaceNombreVacio(t *testing.T) {
	_, err := CreateTAPInterface("")
	if err == nil {
		t.Error("debería fallar con nombre vacío")
	}
}

func TestCreateTAPInterfaceNombreLargo(t *testing.T) {
	_, err := CreateTAPInterface("estenombreesdemasiadolargo")
	if err == nil {
		t.Error("debería fallar con nombre mayor de 15 caracteres")
	}
}

func TestDeleteTAPInterfaceNombreVacio(t *testing.T) {
	err := DeleteTAPInterface("")
	if err == nil {
		t.Error("debería fallar con nombre vacío")
	}
}

func TestCreateTAPInterfaceReal(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("test requiere permisos de root — saltando en CI")
	}

	tap, err := CreateTAPInterface("tap-test0")
	if err != nil {
		t.Fatalf("error creando TAP: %v", err)
	}
	defer DeleteTAPInterface(tap.Name)

	if tap.Name != "tap-test0" {
		t.Errorf("nombre incorrecto: %s", tap.Name)
	}
}
