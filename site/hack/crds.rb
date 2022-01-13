#! /usr/bin/env ruby
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