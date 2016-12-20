#!/usr/bin/env perl

package Watson {
	use strict;
	use warnings;
	use utf8;
	use Text::Shirasu;

	sub new {
		my $class = shift;
		bless +{
			shirasu => Text::Shirasu->new
		}, $class;
	}

	sub normalize { shift->{shirasu}->normalize(@_) }

	sub dbh {
	    my $self = shift;

	    unless (defined $self->{dbh}) {
	        my %params = @_;

	        my $dbi    = $params{db}     // 'Pg';
	        my $dbname = $params{dbname} // 'test';
	        my $host   = $params{host}   // 'localhost';
	        my $port   = $params{port}   // 5432;
	        my $user   = $params{user}   // '';
	        my $pass   = $params{pass}   // '';
	        my $option = $params{option} // {};

	        $self->{dbh} = DBIx::Simple->connect("dbi:${dbi}:dbname=${dbname};host=${host};port=${port};", $user, $pass, $option)
	        	or die DBIx::Simple->error;
	    }

	    return $self->{dbh};
	}
};

use Mojolicious::Lite;
use DBIx::Simple;
use utf8;
use Encode qw/decode_utf8/;
use Text::Shirasu qw/normalize_hyphen normalize_symbols/;
use Data::Dumper;

my $w = Watson->new;

get '/users' => sub {
  my $c = shift;
  my $category = $c->param('category') || '';
  my $start = $c->param('start') || 1;
  my $end = $c->param('end') || 100;
  my $q = "select s.row_id, s.id, s.category from (select row_number() over(order by id asc) as row_id, * from users) as s where s.category = ? and s.row_id between ? and ?";
  my $result = $w->dbh->query($q, $category, $start, $end)->hashes;

  $c->render(json => {
  	users => $result
  });
};

get '/users/:id/tweets' => [id => qr/[0-9]+/] => sub {
  my $c = shift;
  my $user_id = $c->param('id');
  my $is_exist = $w->dbh->query("select id from users where id = ? limit 1", $user_id)->rows;

  if ($is_exist == 0) {
  	return $c->render(json => {
  	  tweets => +[]
    });
  }

  my $start = $c->param('start') || 1;
  my $end = $c->param('end') || 100;
  my $q = "select s.no, s.id, s.user_id, s.text from (select row_number() over(order by id asc) as no, * from tweets) as s where s.user_id = ? and s.no between ? and ?";
  my $result = $w->dbh->query($q, $user_id, $start, $end)->hashes;

  for my $hash ( @$result ) {
  	$hash->{text} = $w->normalize($hash->{text}, qw/
                nfkc
                nfkd
                nfc
                nfd
                alnum_z2h
                space_z2h
                katakana_h2z
                decode_entities
                unify_nl
                unify_whitespaces
                unify_long_spaces
                trim
                old2new_kana
                old2new_kanji
                tab2space
                all_dakuon_normalize
                square2katakana
                circled2kana
                circled2kanji
                decompose_parenthesized_kanji
            /, \&normalize_hyphen, \&normalize_symbols, \&trimgrass, \&trimemoji, \&trimnewline, \&trimurl, \&trimbrackets);
  }

  $c->render(json => {
  	users => $result
  });
};

sub trimbrackets {
	local $_ = shift;
	s/[\x{3008}\x{300a}\x{3010}].*[\x{3009}\x{300b}\x{3011}]//g;
	$_;
}

sub trimgrass {
	local $_ = shift;
	s/w+//g;
	$_;
}

sub trimemoji {
	local $_ = shift;
	s/[\x{2605}-\x{27bf}]//g;
	$_;
}

sub trimnewline {
	local $_ = shift;
	s/\s+//g;
	$_;
}

sub trimurl {
	local $_ = shift;
	s{(http.*?://([^\s)\"](?!ttp:))+)}{}g;
	$_;
}

app->start;
