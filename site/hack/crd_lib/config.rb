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

class Config
  class Fields
    attr_reader :force_hide_descriptions,
      :max_depth

    def initialize(fields_hash, force_hide_descriptions = false, max_depth = 99)
      @fields_hash = fields_hash
      @fields_memo = {}
      @force_hide_descriptions = force_hide_descriptions
      @max_depth = max_depth
    end

    def [](field_name)
      @fields_memo[field_name] ||= Field.new(field_name, @fields_hash[field_name] || {}, force_hide_descriptions, max_depth)
    end
  end

  class Field
    attr_accessor :name,
      :fields,
      :append_description,
      :max_depth

    def initialize(field_name, field_hash, force_hide_descriptions = false, max_depth = 99)
      @max_depth = [max_depth, (field_hash["maxDepth"] || 99)].min
      @name = field_name
      @hide = field_hash["hide"] || (max_depth < 0)
      @append_description = field_hash["appendDescription"]
      @hide_description = force_hide_descriptions || field_hash["hideDescription"]

      fields_hash = field_hash["fields"] || {}
      hide_children = force_hide_descriptions || field_hash["hideChildDescriptions"]
      @fields = Fields.new(fields_hash, hide_children, @max_depth - 1)
    end

    def hide?
      @hide
    end

    def append_description?
      !@append_description.nil?
    end

    def hide_description?
      @hide_description
    end

    def hide_child_descriptions?
      @hide_child_descriptions
    end
  end

  attr_reader :version, :fields, :comment_wrap_at

  def initialize(config_hash)
    @version = config_hash.dig("version") || "v1"
    @comment_wrap_at = (config_hash.dig("commentWrapAt") || 64).to_i
    @fields = Fields.new(config_hash["fields"]) if config_hash.has_key?("fields")
  end

  def apply!(crd)
    schema_versions = crd["spec"]["versions"]
    schema_version = schema_versions.find { |v| v["name"] == version }
    schema = schema_version["schema"]["openAPIV3Schema"]
    apply_object(fields, schema)
  end

  private

  def apply_field(field, obj)
    property = obj["properties"][field.name]

    if field.append_description?
      property["description"] ||= ""
      property["description"] += "\n#{field.append_description}"
    elsif field.hide_description?
      property.delete("description")
    end

    if property
      if field.hide?
        obj["properties"].delete(field.name)
      else
        case property["type"]
        when "object"
          apply_object(field.fields, property)
        when "array"
          apply_array(field.fields, property)
        end
      end
    else
      puts "Error: Config field `#{field.name}` not found"
    end
  end

  def apply_object(fields, obj)
    (obj["properties"] || {}).each do |name, property|
      apply_field(fields[name], obj)
    end
  end

  def apply_array(fields, obj)
    (obj["items"]["properties"] || {}).each do |name, property|
      apply_field(fields[name], obj["items"])
    end
  end

end
