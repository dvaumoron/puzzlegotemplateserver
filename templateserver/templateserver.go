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
	"text/template"
	"time"

	"github.com/dvaumoron/partrenderer"
	pb "github.com/dvaumoron/puzzletemplateservice"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

const TemplateKey = "puzzleTemplate"

var errInternal = errors.New("internal service error")

// server is used to implement puzzletemplateservice.TemplateServer
type server struct {
	pb.UnimplementedTemplateServer
	renderer partrenderer.PartRenderer
	messages map[string]map[string]string
	logger   *otelzap.Logger
}

func New(componentsPath string, viewsPath string, sourceFormat string, messages map[string]map[string]string, logger *otelzap.Logger) pb.TemplateServer {
	renderer := load(componentsPath, viewsPath, sourceFormat, logger)
	return server{renderer: renderer, messages: messages, logger: logger}
}

func (s server) Render(ctx context.Context, request *pb.RenderRequest) (*pb.Rendered, error) {
	logger := s.logger.Ctx(ctx)

	var data map[string]any
	err := json.Unmarshal(request.Data, &data)
	if err != nil {
		logger.Error("Failed during JSON parsing", zap.Error(err))
		return nil, errInternal
	}
	data["Messages"] = s.messages[asString(data["lang"])]
	var content bytes.Buffer
	if err = s.renderer.ExecuteTemplate(&content, request.TemplateName, data); err != nil {
		logger.Error("Failed during go template call", zap.Error(err))
		return nil, errInternal
	}
	return &pb.Rendered{Content: content.Bytes()}, nil
}

func load(componentsPath string, viewsPath string, sourceFormat string, logger *otelzap.Logger) partrenderer.PartRenderer {
	customFuncs := template.FuncMap{"date": func(value string, targetFormat string) string {
		if sourceFormat == targetFormat {
			return value
		}
		date, err := time.Parse(sourceFormat, value)
		if err != nil {
			return value
		}
		return date.Format(targetFormat)
	}}

	renderer, err := partrenderer.MakePartRenderer(componentsPath, viewsPath, ".html", customFuncs)
	if err != nil {
		logger.Fatal("Failed to load templates", zap.Error(err))
	}
	return renderer
}

func asString(value any) string {
	s, _ := value.(string)
	return s
}
