require_relative './string'

OPTIONAL_STRING = "+optional"

class Writer
  attr_accessor :out
  attr_accessor :wrap

  def initialize(out = $stdout, wrap = 64, optional_comment = OPTIONAL_STRING)
    @indent = 0
    @out = out
    @wrap = wrap
    @array_mode = false
    @optional_comment = optional_comment
  end

  def indent(array_mode = false, &block)
    @array_mode = array_mode
    if block_given?
      @indent += 1
      block.call(self)
      @indent -= 1
    else
      @indent += 1
    end
  end

  def outdent(&block)
    @array_mode = false # always
    @indent -= 1
  end

  def puts(msg = "")
    @out.puts("#{gen_indent}#{msg}")
  end

  def comment(comment, optional)
    return if comment.nil?
    @out.puts unless @array_mode
    comment_lines = comment.word_wrap(@wrap - (@indent * 2))
    comment_lines.split("\n").each do |line|
      @out.puts("#{gen_indent}# #{line}")
    end
    @out.puts("#{gen_indent}# #{@optional_comment}") if optional
  end

  private

  def gen_indent
    return "  " * @indent unless @array_mode
    @array_mode = false
    ("  " * (@indent - 1)) + "- "
  end
end
