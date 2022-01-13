class String
  def word_wrap(col_width = 80)
    self.dup.word_wrap!(col_width)
  end

  def word_wrap!(col_width = 80)
    # commented this out as it forces long lines split (URLS get smashed)
    # self.gsub!( /(\S{#{col_width}})(?=\S)/, '\1 ' )
    self.gsub!(/(.{1,#{col_width}})(?:\s+|$)/, "\\1\n")
    self
  end
end
