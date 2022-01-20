#! /usr/bin/env ruby
# Copyright 2021 VMware
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

require 'yaml'
require 'fileutils'
require_relative 'crd_lib/config'
require_relative 'crd_lib/spec'

CONFIG_DIR = File.join(__dir__, "crds")
CRD_DIR = File.join(__dir__, "..", "..", "config", "crd", "bases")
OUT_DIR = File.join(__dir__, "..", "content", "docs", "development", "crds")

def main
  FileUtils.mkdir_p OUT_DIR
  config_files = Dir.glob(File.join(CONFIG_DIR, '*.yaml'))

  config_files.each do |filename|
    puts "Processing: #{filename}"

    # Load Config
    config_object = YAML.load_file(filename)
    config = Config.new(config_object)

    # Load Input CRD Spec
    input_filename = File.join(CRD_DIR, File.basename(filename))
    puts "\tSource CRD: #{File.absolute_path(input_filename)}"

    crd = YAML.load_file(input_filename)

    config.apply!(crd)

    spec = Spec.new(crd["spec"])

    # Create/Open Target Spec Yaml
    output_filename = File.join(OUT_DIR, File.basename(filename))
    puts "\tTarget Resource Spec: #{File.absolute_path(output_filename)}"

    output_file = File.open(output_filename, "w")

    writer = Writer.new(output_file, config.comment_wrap_at)
    spec.versions[config.version].to_y(writer)

    output_file.close

  end
end

main