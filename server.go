/*
 *
 * Copyright 2023 puzzlegotemplateserver authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	_ "embed"
	"os"
	"strings"

	"github.com/dvaumoron/puzzlegotemplateserver/templateserver"
	grpcserver "github.com/dvaumoron/puzzlegrpcserver"
	locale "github.com/dvaumoron/puzzlelocaleloader"
	pb "github.com/dvaumoron/puzzletemplateservice"
	"go.uber.org/zap"
)

//go:embed version.txt
var version string

func main() {
	s := grpcserver.Make(templateserver.TemplateKey, version)

	confLangs := strings.Split(os.Getenv("AVAILABLE_LOCALES"), ",")
	allLang := make([]string, 0, len(confLangs))
	for _, lang := range confLangs {
		allLang = append(allLang, strings.TrimSpace(lang))
	}

	componentsPath := os.Getenv("COMPONENTS_PATH")
	viewsPath := os.Getenv("VIEWS_PATH")
	localesPath := os.Getenv("LOCALES_PATH")
	messages, err := locale.Load(localesPath, allLang)
	if err != nil {
		s.Logger.Fatal("Failed to load locale files", zap.Error(err))
	}
	sourceFormat := os.Getenv("DATE_FORMAT")

	pb.RegisterTemplateServer(s, templateserver.New(componentsPath, viewsPath, sourceFormat, messages, s.Logger))
	s.Start()
}
