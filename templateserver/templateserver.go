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
package templateserver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"

	pb "github.com/dvaumoron/puzzletemplateservice"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

const TemplateKey = "puzzleTemplate"

var errInternal = errors.New("internal service error")

// server is used to implement puzzletemplateservice.TemplateServer
type server struct {
	pb.UnimplementedTemplateServer
	templates *template.Template
	logger    *otelzap.Logger
}

func New(templatesPath string, logger *otelzap.Logger) pb.TemplateServer {
	tmpl := load(templatesPath, logger)
	return server{templates: tmpl, logger: logger}
}

func (s server) Render(ctx context.Context, request *pb.RenderRequest) (*pb.Rendered, error) {
	logger := s.logger.Ctx(ctx)

	var data map[string]any
	err := json.Unmarshal(request.Data, &data)
	if err != nil {
		logger.Error("Failed during JSON parsing", zap.Error(err))
		return nil, errInternal
	}
	var content bytes.Buffer
	if err = s.templates.ExecuteTemplate(&content, request.TemplateName, data); err != nil {
		logger.Error("Failed during go template call", zap.Error(err))
		return nil, errInternal
	}
	return &pb.Rendered{Content: content.Bytes()}, nil
}

func load(templatesPath string, logger *otelzap.Logger) *template.Template {
	templatesPath, err := filepath.Abs(templatesPath)
	if err != nil {
		logger.Fatal("Wrong templatesPath", zap.Error(err))
	}
	if last := len(templatesPath) - 1; templatesPath[last] != '/' {
		templatesPath += "/"
	}

	tmpl := template.New("")
	inSize := len(templatesPath)
	err = filepath.WalkDir(templatesPath, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			name := path[inSize:]
			if name[len(name)-5:] == ".html" {
				var data []byte
				data, err = os.ReadFile(path)
				if err == nil {
					_, err = tmpl.New(name).Parse(string(data))
				}
			}
		}
		return err
	})

	if err != nil {
		logger.Fatal("Failed to load templates", zap.Error(err))
	}
	return tmpl
}
