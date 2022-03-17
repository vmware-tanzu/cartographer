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

require_relative 'writer'

class Spec

  class Versions
    def initialize(versions_hash, group, kind)
      @group = group
      @kind = kind
      @versions_hash = versions_hash
      @versions_memo = {}
    end

    def [](name)
      @versions_memo[name] ||= Version.new(@group, name, @kind, @versions_hash[name])
    end
  end

  class Version
    def self.node_factory(node_hash, name = "", required = false)
      case node_hash["type"]
      when "object"
        ObjectNode.new(node_hash, name, required)
      when "array"
        ArrayNode.new(node_hash, name, required)
      else
        ScalarNode.new(node_hash, name, required)
      end
    end

    class ItemsNode
      def initialize(items)
        @items = ObjectNode.new(items)
      end

      def to_y(writer)
        writer.indent(true) { @items.to_y(writer) }
      end
    end

    class PropertiesNode
      include Enumerable

      def initialize(properties_hash, required_hash)
        @properties_hash = properties_hash || {}
        @required_hash = required_hash || {}
        @properties_memo = {}
      end

      def required?(property_name)
        @required_hash.include?(property_name)
      end

      def [](name)
        @properties_memo[name] ||= Version.node_factory(@properties_hash[name], name, required?(name))
      end

      def each(&block)
        if block_given?
          @properties_hash.each do |key, value|
            block.call(self[key])
          end
        else
          to_enum(:each)
        end
      end

      def empty?
        @properties_hash.length == 0
      end
    end

    class ScalarNode
      attr_reader :name, :description

      def initialize(node_hash, name = "", required = false)
        @name = name
        @required = required
        @description = node_hash["description"]
        @node_hash = node_hash
      end

      def required?
        @required
      end

      def to_y(writer)
        writer.comment(description, !required?)
        value = @node_hash.dig("value") || "<#{type_string}>"
        writer.puts "#{name}: #{value}"
      end

      private

      def any_of_type_string
        any_of = @node_hash["anyOf"] || []
        return nil if any_of.empty?
        "[" + any_of.map { |of| of.dig("type") }.join("|") + "]"
      end

      def enum_type_string
        return unless @node_hash.has_key?("enum")
        "[#{@node_hash["enum"].join("|")}]"
      end

      def type_string
        enum_type_string || @node_hash["type"] || any_of_type_string || "any"
      end

    end

    class ObjectNode
      attr_reader :name,
        :properties,
        :description

      def initialize(node_hash, name = "", required = false)
        @name = name
        @required = required
        @description = node_hash["description"]
        @properties = PropertiesNode.new(node_hash.dig("properties"), node_hash.dig("required"))
      end

      def required?
        @required
      end

      def named?
        name != ""
      end

      def to_y(writer)
        writer.comment(description, !required?)
        if named? && properties.empty?
          writer.puts "#{name}: {}"
          return
        end

        if named?
          writer.puts "#{name}:"
          writer.indent
        end
        properties.each do |property|
          property.to_y(writer)
        end
        if named?
          writer.outdent
        end
      end
    end

    class ArrayNode
      attr_reader :name, :items, :description, :scalar_item_type

      def initialize(node_hash, name = "", required = false)
        @name = name
        @description = node_hash["description"]
        @required = required

        if node_hash["items"]["type"] == "object"
          @items = ItemsNode.new(node_hash.dig("items"))
        else
          @scalar_item_type = node_hash["items"]["type"]
        end
      end

      def required?
        @required
      end

      def to_y(writer)
        writer.comment(description, !required?)
        if @items
          writer.puts "#{name}:"
          writer.indent { @items.to_y(writer) }
        elsif @scalar_item_type
          writer.puts "#{name}: [ <#{scalar_item_type}> ]"
        end
      end
    end

    attr_reader :group, :version, :kind, :schema

    def initialize(group, version, kind, version_hash)
      @group = group
      @version = version
      @kind = kind
      schema_hash = version_hash.dig("schema", "openAPIV3Schema") || {}
      schema_hash["properties"]["apiVersion"]["value"] = "#{group}/#{version}"
      schema_hash["properties"]["kind"]["value"] = kind
      @schema = Version.node_factory(schema_hash)
    end

    def to_y(writer)
      writer.puts "---"
      @schema.to_y(writer)
    end

  end

  attr_reader :versions

  def initialize(spec_hash)
    @group = spec_hash.dig("group")
    @kind = spec_hash.dig("names", "kind")
    versions_array = spec_hash.dig("versions")
    versions_hash = versions_array.inject({}) { |hash, version| hash[version["name"]] = version; hash }
    @versions = Versions.new(versions_hash, @group, @kind)
  end
end