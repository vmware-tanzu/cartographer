class String
  def word_wrap(col_width = 80)
    self.dup.word_wrap!(col_width)
  end

  def word_wrap!(col_width = 80)
    self.gsub!(/(.{1,#{col_width}})(?:\s+|$)/, "\\1\n")
    self
  end
end
