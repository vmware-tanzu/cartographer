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
