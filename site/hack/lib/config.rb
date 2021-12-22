class Config

  class Fields
    def initialize(fields_hash)
      @fields = {}
      fields_hash.each do |field_name, field_hash|
        @fields[field_name] = Field.new(field_name, field_hash)
      end
    end

    def [](field_name)
      @fields[field_name]
    end
  end

  class Field
    attr_accessor :name, :fields

    def initialize(field_name, field_hash)
      @name = field_name
      @fields = Fields.new(field_hash["fields"]) if field_hash.has_key?("fields")
    end
  end

  attr_reader :version, :fields, :comment_wrap_at

  def initialize(config_hash)
    @version = config_hash.dig("version") || "v1"
    @comment_wrap_at = (config_hash.dig("commentWrapAt") || 64).to_i
    @fields = Fields.new(config_hash["fields"]) if config_hash.has_key?("fields")
  end
end
